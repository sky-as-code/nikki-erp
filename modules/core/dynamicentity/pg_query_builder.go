package dynamicentity

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
)

// QueryBuilder defines the interface for SQL query generation.
type QueryBuilder interface {
	SqlCreateTable(dbEntity *dschema.DbEntity) (string, error)
	SqlSelectGraph(dbEntity *dschema.DbEntity, graph dschema.SearchGraph, columns []string) (string, error)
	SqlInsertMap(dbEntity *dschema.DbEntity, data dmodel.EntityMap) (string, error)
	SqlInsertBulkMaps(dbEntity *dschema.DbEntity, rows []dmodel.EntityMap) (string, error)
	SqlInsertStruct(dbEntity *dschema.DbEntity, payload any) (string, error)
	SqlInsertBulkStructs(dbEntity *dschema.DbEntity, items []any) (string, error)
	SqlUpdateByPkMap(dbEntity *dschema.DbEntity, data dmodel.EntityMap) (string, error)
	SqlUpdateByPkStruct(dbEntity *dschema.DbEntity, payload any) (string, error)
	SqlDeleteEqualStruct(dbEntity *dschema.DbEntity, filters dmodel.EntityMap) (string, error)
}

// PgQueryBuilder implements QueryBuilder for PostgreSQL.
type PgQueryBuilder struct {
}

// NewPgQueryBuilder creates a new PgQueryBuilder.
func NewPgQueryBuilder() *PgQueryBuilder {
	return &PgQueryBuilder{}
}

func (this *PgQueryBuilder) SqlCreateTable(dbEntity *dschema.DbEntity) (string, error) {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(dbEntity.Name())
	for _, col := range dbEntity.Columns() {
		builder.Define(col.Name, col.Type, col.Nullable)
	}

	if keys := dbEntity.KeyColumns(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(keys), ", ")))
	}

	for _, unique := range dbEntity.UniqueKeys() {
		if len(unique) == 0 {
			continue
		}
		builder.Define("UNIQUE", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(unique), ", ")))
	}

	sql, _ := builder.Build()
	return sql, nil
}

func (this *PgQueryBuilder) SqlSelectGraph(dbEntity *dschema.DbEntity, graph dschema.SearchGraph, columns []string) (string, error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 4)

	if len(columns) == 0 {
		mods = append(mods, sm.Columns(psql.Raw("*")))
	} else {
		columnExprs := make([]any, len(columns))
		for i, col := range columns {
			if _, ok := dbEntity.Column(col); !ok {
				return "", fmt.Errorf("unknown column '%s'", col)
			}
			columnExprs[i] = psql.Quote(col)
		}
		mods = append(mods, sm.Columns(columnExprs...))
	}

	mods = append(mods, sm.From(this.tableExpression(dbEntity)))

	predicate, ok, err := this.graphExpression(dbEntity, graph.Condition, graph.And, graph.Or)
	if err != nil {
		return "", err
	}
	if ok {
		mods = append(mods, sm.Where(predicate))
	}

	orderMods, err := this.orderMods(dbEntity, graph.Order)
	if err != nil {
		return "", err
	}
	mods = append(mods, orderMods...)

	return this.buildSQL(psql.Select(mods...))
}

func (this *PgQueryBuilder) tableExpression(dbEntity *dschema.DbEntity) any {
	tableName := dbEntity.TableName()
	if tableName == "" {
		tableName = dbEntity.Name()
	}
	parts := strings.Split(tableName, ".")
	return psql.Quote(parts...)
}

func (this *PgQueryBuilder) SqlInsertMap(dbEntity *dschema.DbEntity, data dmodel.EntityMap) (string, error) {
	return this.insertFromMaps(dbEntity, []dmodel.EntityMap{data})
}

func (this *PgQueryBuilder) SqlInsertBulkMaps(dbEntity *dschema.DbEntity, rows []dmodel.EntityMap) (string, error) {
	return this.insertFromMaps(dbEntity, rows)
}

func (this *PgQueryBuilder) SqlInsertStruct(dbEntity *dschema.DbEntity, payload any) (string, error) {
	m, err := this.structToMap(payload)
	if err != nil {
		return "", err
	}
	return this.SqlInsertMap(dbEntity, m)
}

func (this *PgQueryBuilder) SqlInsertBulkStructs(dbEntity *dschema.DbEntity, items []any) (string, error) {
	maps, err := this.structSliceToMaps(items)
	if err != nil {
		return "", err
	}
	return this.SqlInsertBulkMaps(dbEntity, maps)
}

func (this *PgQueryBuilder) SqlUpdateByPkMap(dbEntity *dschema.DbEntity, data dmodel.EntityMap) (string, error) {
	return this.updateFromMap(dbEntity, data)
}

func (this *PgQueryBuilder) SqlUpdateByPkStruct(dbEntity *dschema.DbEntity, payload any) (string, error) {
	m, err := this.structToMap(payload)
	if err != nil {
		return "", err
	}
	return this.updateFromMap(dbEntity, m)
}

func (this *PgQueryBuilder) SqlDeleteEqualStruct(dbEntity *dschema.DbEntity, filters dmodel.EntityMap) (string, error) {
	if len(filters) == 0 {
		return "", errors.New("no filters provided")
	}
	if err := this.ensureTenantKeyInMap(dbEntity, filters); err != nil {
		return "", err
	}

	row, err := this.rowFromMap(dbEntity, filters, nil)
	if err != nil {
		return "", err
	}
	if len(row.columns) == 0 {
		return "", errors.New("no filters provided")
	}

	mods := []bob.Mod[*dialect.DeleteQuery]{dm.From(this.tableExpression(dbEntity))}
	for i, col := range row.columns {
		mods = append(mods, dm.Where(psql.Quote(col).EQ(psql.Arg(row.values[i]))))
	}

	return this.buildSQL(psql.Delete(mods...))
}

func (this *PgQueryBuilder) ensureTenantKeyInMap(dbEntity *dschema.DbEntity, values dmodel.EntityMap) error {
	key := dbEntity.TenantKey()
	if key == "" {
		return nil
	}
	if _, ok := values[key]; !ok {
		return fmt.Errorf("missing tenant key '%s'", key)
	}
	return nil
}

func (this *PgQueryBuilder) insertFromMaps(dbEntity *dschema.DbEntity, rows []dmodel.EntityMap) (string, error) {
	prepared, err := this.rowsFromMaps(dbEntity, rows, nil)
	if err != nil {
		return "", err
	}
	query, err := this.insertQuery(dbEntity, prepared)
	if err != nil {
		return "", err
	}
	return this.buildSQL(query)
}

func (this *PgQueryBuilder) updateFromMap(dbEntity *dschema.DbEntity, data dmodel.EntityMap) (string, error) {
	if len(dbEntity.PrimaryKeys()) == 0 {
		return "", errors.New("entity has no primary keys")
	}
	if err := this.ensureTenantKeyInMap(dbEntity, data); err != nil {
		return "", err
	}

	target, err := this.rowFromMap(dbEntity, data, func(name string) bool {
		return !dbEntity.IsPrimaryKey(name) && !dbEntity.IsTenantKey(name)
	})
	if err != nil {
		return "", err
	}
	if len(target.columns) == 0 {
		return "", errors.New("no updatable columns provided")
	}

	lookup, err := this.rowForKeys(dbEntity, data, dbEntity.KeyColumns())
	if err != nil {
		return "", err
	}
	mods := []bob.Mod[*dialect.UpdateQuery]{um.Table(this.tableExpression(dbEntity))}
	for i, col := range target.columns {
		mods = append(mods, um.SetCol(col).ToArg(target.values[i]))
	}
	for i, col := range lookup.columns {
		mods = append(mods, um.Where(psql.Quote(col).EQ(psql.Arg(lookup.values[i]))))
	}

	return this.buildSQL(psql.Update(mods...))
}

type rowData struct {
	columns []string
	values  []any
}

func (this *PgQueryBuilder) rowsFromMaps(entity *dschema.DbEntity, rows []dmodel.EntityMap, filter func(string) bool) ([]rowData, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		if err := this.ensureTenantKeyInMap(entity, row); err != nil {
			return nil, err
		}
		item, err := this.rowFromMap(entity, row, filter)
		if err != nil {
			return nil, err
		}
		if len(item.columns) == 0 {
			return nil, errors.New("no columns provided")
		}
		if index == 0 {
			reference = item.columns
		} else if !equalStrings(reference, item.columns) {
			return nil, fmt.Errorf("row %d column mismatch", index)
		}
		prepared[index] = item
	}

	return prepared, nil
}

func (this *PgQueryBuilder) rowFromMap(entity *dschema.DbEntity, values dmodel.EntityMap, include func(string) bool) (rowData, error) {
	includeFn := include
	if includeFn == nil {
		includeFn = func(string) bool { return true }
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if includeFn(key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	result := rowData{
		columns: keys,
		values:  make([]any, len(keys)),
	}
	for i, key := range keys {
		col, ok := entity.Column(key)
		if !ok {
			return rowData{}, fmt.Errorf("unknown column '%s'", key)
		}
		converted, err := this.convertValue(col, values[key])
		if err != nil {
			return rowData{}, fmt.Errorf("invalid value for column '%s': %w", key, err)
		}
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) rowForKeys(entity *dschema.DbEntity, values dmodel.EntityMap, keys []string) (rowData, error) {
	result := rowData{
		columns: make([]string, len(keys)),
		values:  make([]any, len(keys)),
	}

	for i, key := range keys {
		col, ok := entity.Column(key)
		if !ok {
			return rowData{}, fmt.Errorf("unknown key column '%s'", key)
		}
		raw, ok := values[key]
		if !ok {
			return rowData{}, fmt.Errorf("missing key '%s'", key)
		}
		converted, err := this.convertValue(col, raw)
		if err != nil {
			return rowData{}, fmt.Errorf("invalid value for key '%s': %w", key, err)
		}
		result.columns[i] = key
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) insertQuery(entity *dschema.DbEntity, rows []rowData) (bob.Query, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}
	columns := rows[0].columns

	mods := []bob.Mod[*dialect.InsertQuery]{im.Into(this.tableExpression(entity), columns...)}
	for _, row := range rows {
		mods = append(mods, im.Values(psql.Arg(row.values...)))
	}

	return psql.Insert(mods...), nil
}

func (this *PgQueryBuilder) structToMap(payload any) (dmodel.EntityMap, error) {
	if payload == nil {
		return nil, errors.New("nil payload")
	}
	value := reflect.ValueOf(payload)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, errors.New("nil payload")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", payload)
	}
	return this.exportStructFields(value)
}

func (this *PgQueryBuilder) exportStructFields(value reflect.Value) (dmodel.EntityMap, error) {
	result := make(dmodel.EntityMap)
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag := strings.Split(field.Tag.Get(dschema.SchemaStructTag), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		result[tag] = value.Field(i).Interface()
	}
	if len(result) == 0 {
		return nil, errors.New("struct has no db-tagged fields")
	}
	return result, nil
}

func (this *PgQueryBuilder) structSliceToMaps(items []any) ([]dmodel.EntityMap, error) {
	if len(items) == 0 {
		return nil, errors.New("no payload provided")
	}
	result := make([]dmodel.EntityMap, len(items))
	for i, item := range items {
		m, err := this.structToMap(item)
		if err != nil {
			return nil, err
		}
		result[i] = m
	}
	return result, nil
}

func (this *PgQueryBuilder) buildSQL(query bob.Query) (string, error) {
	sql, args, err := bob.Build(context.Background(), query)
	if err != nil {
		return "", err
	}
	if len(args) == 0 {
		return sql, nil
	}

	iface := make([]interface{}, len(args))
	for i, arg := range args {
		iface[i] = arg
	}
	return sqlbuilder.PostgreSQL.Interpolate(sql, iface)
}

func (this *PgQueryBuilder) graphExpression(
	entity *dschema.DbEntity,
	condition *dschema.Condition,
	and []dschema.SearchNode,
	or []dschema.SearchNode,
) (expr bob.Expression, ok bool, err error) {
	switch {
	case condition != nil:
		expr, err = this.conditionExpression(entity, *condition)
		return expr, err == nil, err
	case len(and) > 0:
		return this.combineNodes(entity, and, psql.And)
	case len(or) > 0:
		return this.combineNodes(entity, or, psql.Or)
	default:
		return nil, false, nil
	}
}

func (this *PgQueryBuilder) combineNodes(
	entity *dschema.DbEntity,
	nodes []dschema.SearchNode,
	join func(...bob.Expression) psql.Expression,
) (expr bob.Expression, ok bool, err error) {
	expressions := make([]bob.Expression, 0, len(nodes))
	for _, node := range nodes {
		predicate, predicateOk, predicateErr := this.graphExpression(entity, node.Condition, node.And, node.Or)
		if predicateErr != nil {
			return nil, false, predicateErr
		}
		if predicateOk {
			expressions = append(expressions, predicate)
		}
	}
	if len(expressions) == 0 {
		return nil, false, nil
	}
	return join(expressions...), true, nil
}

func (this *PgQueryBuilder) conditionExpression(entity *dschema.DbEntity, cond dschema.Condition) (bob.Expression, error) {
	col, expr, err := this.prepareCondition(entity, cond.Field())
	if err != nil {
		return nil, err
	}

	switch cond.Operator() {
	case dschema.Equals, dschema.NotEquals, dschema.GreaterThan, dschema.GreaterEqual, dschema.LessThan, dschema.LessEqual:
		return this.comparisonPredicate(expr, col, cond.Operator(), cond.Value())
	case dschema.In, dschema.NotIn:
		return this.collectionPredicate(expr, col, cond.Operator(), cond.Values())
	case dschema.Contains, dschema.NotContains, dschema.StartsWith, dschema.NotStartsWith, dschema.EndsWith, dschema.NotEndsWith:
		return this.stringPredicate(expr, col, cond.Operator(), cond.Value())
	case dschema.IsSet, dschema.IsNotSet:
		return this.nullPredicate(expr, cond.Operator()), nil
	default:
		return nil, fmt.Errorf("unsupported operator '%s'", cond.Operator())
	}
}

func (this *PgQueryBuilder) prepareCondition(entity *dschema.DbEntity, field string) (dschema.Column, psql.Expression, error) {
	if strings.Contains(field, ".") {
		return dschema.Column{}, psql.Expression{}, fmt.Errorf("nested fields not supported: %s", field)
	}
	col, ok := entity.Column(field)
	if !ok {
		return dschema.Column{}, psql.Expression{}, fmt.Errorf("unknown column '%s'", field)
	}
	return col, psql.Quote(col.Name), nil
}

func (this *PgQueryBuilder) comparisonPredicate(expr psql.Expression, col dschema.Column, op dschema.Operator, value any) (bob.Expression, error) {
	converted, err := this.convertValue(col, value)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted)

	switch op {
	case dschema.Equals:
		return expr.EQ(arg), nil
	case dschema.NotEquals:
		return expr.NE(arg), nil
	case dschema.GreaterThan:
		return expr.GT(arg), nil
	case dschema.GreaterEqual:
		return expr.GTE(arg), nil
	case dschema.LessThan:
		return expr.LT(arg), nil
	case dschema.LessEqual:
		return expr.LTE(arg), nil
	default:
		return nil, fmt.Errorf("unsupported comparison operator '%s'", op)
	}
}

func (this *PgQueryBuilder) collectionPredicate(expr psql.Expression, col dschema.Column, op dschema.Operator, values []any) (bob.Expression, error) {
	converted, err := this.convertValues(col, values)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted...)

	if op == dschema.In {
		return expr.In(arg), nil
	}
	if op == dschema.NotIn {
		return expr.NotIn(arg), nil
	}
	return nil, fmt.Errorf("unsupported collection operator '%s'", op)
}

func (this *PgQueryBuilder) stringPredicate(expr psql.Expression, col dschema.Column, op dschema.Operator, value any) (bob.Expression, error) {
	if columnCategoryFor(col.Type) != columnString {
		return nil, fmt.Errorf("operator '%s' requires string column '%s'", op, col.Name)
	}

	converted, err := this.convertValue(col, value)
	if err != nil {
		return nil, err
	}
	pattern := stringPattern(fmt.Sprint(converted), op)
	if pattern == "" {
		return nil, fmt.Errorf("unsupported string operator '%s'", op)
	}

	like := expr.ILike(psql.Arg(pattern))
	if strings.HasPrefix(string(op), "!") {
		return psql.Not(like), nil
	}
	return like, nil
}

func (this *PgQueryBuilder) nullPredicate(expr psql.Expression, op dschema.Operator) bob.Expression {
	if op == dschema.IsSet {
		return expr.IsNotNull()
	}
	return expr.IsNull()
}

func (this *PgQueryBuilder) convertValues(col dschema.Column, values []any) ([]any, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("operator requires at least one value for column '%s'", col.Name)
	}
	converted := make([]any, len(values))
	for i, value := range values {
		next, err := this.convertValue(col, value)
		if err != nil {
			return nil, err
		}
		converted[i] = next
	}
	return converted, nil
}

func (this *PgQueryBuilder) orderMods(entity *dschema.DbEntity, order dschema.SearchOrder) ([]bob.Mod[*dialect.SelectQuery], error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		field := item[0]
		if strings.Contains(field, ".") {
			return nil, fmt.Errorf("nested order not supported: %s", field)
		}
		if _, ok := entity.Column(field); !ok {
			return nil, fmt.Errorf("unknown order column '%s'", field)
		}
		mod := sm.OrderBy(psql.Quote(field))
		if item.Direction() == dschema.Desc {
			mod = mod.Desc()
		} else {
			mod = mod.Asc()
		}
		mods = append(mods, mod)
	}
	return mods, nil
}

func (this *PgQueryBuilder) convertValue(col dschema.Column, value any) (any, error) {
	if value == nil {
		if col.IsNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("column '%s' does not allow NULL", col.Name)
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if col.IsNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("column '%s' does not allow NULL", col.Name)
	}
	if !valueAllowed(columnCategoryFor(col.Type), v) {
		return nil, fmt.Errorf("column '%s': incompatible value type %T", col.Name, v.Interface())
	}
	return v.Interface(), nil
}

type columnCategory int

const (
	columnUnknown columnCategory = iota
	columnString
	columnBool
	columnInt
	columnNumeric
	columnTime
	columnJSON
)

func isIntKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}

func isFloatKind(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}

func isBigNumber(value any) bool {
	switch value.(type) {
	case big.Int, *big.Int, big.Float, *big.Float, big.Rat, *big.Rat:
		return true
	default:
		return false
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func escapeIdentifiers(columns []string) []string {
	escaped := make([]string, len(columns))
	for i, col := range columns {
		escaped[i] = sqlbuilder.Escape(col)
	}
	return escaped
}

func stringPattern(value string, op dschema.Operator) string {
	switch op {
	case dschema.Contains, dschema.NotContains:
		return "%" + value + "%"
	case dschema.StartsWith, dschema.NotStartsWith:
		return value + "%"
	case dschema.EndsWith, dschema.NotEndsWith:
		return "%" + value
	default:
		return ""
	}
}

func columnCategoryFor(t string) columnCategory {
	typ := strings.TrimSpace(strings.ToLower(t))
	switch {
	case strings.Contains(typ, "json"):
		return columnJSON
	case strings.Contains(typ, "bool"):
		return columnBool
	case strings.Contains(typ, "int"):
		return columnInt
	case strings.Contains(typ, "numeric"), strings.Contains(typ, "decimal"), strings.Contains(typ, "float"):
		return columnNumeric
	case strings.Contains(typ, "timestamp"), strings.Contains(typ, "timestamptz"), strings.Contains(typ, "date"), strings.Contains(typ, "time"):
		return columnTime
	case strings.Contains(typ, "char"), strings.Contains(typ, "text"), strings.Contains(typ, "uuid"):
		return columnString
	default:
		return columnUnknown
	}
}

func unwrapValue(v reflect.Value) (reflect.Value, bool) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
	return v, true
}

var timeType = reflect.TypeOf(time.Time{})

func valueAllowed(cat columnCategory, v reflect.Value) bool {
	switch cat {
	case columnString:
		return v.Kind() == reflect.String
	case columnBool:
		return v.Kind() == reflect.Bool
	case columnInt:
		return isIntKind(v.Kind()) || isUintKind(v.Kind())
	case columnNumeric:
		return isIntKind(v.Kind()) || isUintKind(v.Kind()) || isFloatKind(v.Kind()) || isBigNumber(v.Interface())
	case columnTime:
		return v.Type() == timeType
	case columnJSON:
		return v.Kind() == reflect.Map
	default:
		return true
	}
}
