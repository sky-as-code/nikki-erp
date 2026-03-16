package orm

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"slices"
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

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
)

type QueryBuilder interface {
	SqlCreateTable(schema *schema.EntitySchema) (string, error)
	SqlSelectGraph(schema *schema.EntitySchema, graph schema.SearchGraph, columns []string) (string, error)
	SqlInsert(schema *schema.EntitySchema, data schema.DynamicEntity) (string, error)
	SqlInsertBulk(schema *schema.EntitySchema, rows []schema.DynamicEntity) (string, error)
	SqlUpdateByPk(schema *schema.EntitySchema, data schema.DynamicEntity) (string, error)
	SqlDeleteEqual(schema *schema.EntitySchema, filters schema.DynamicEntity) (string, error)
	// SqlCheckUniqueCollisions builds SQL that returns 1 per row where the unique key has a collision, else 0.
	// Input: uniqueKeysToCheck - subset of schema.AllUniques() where data has all values (no nil).
	// Returns (sql, args, nil) for execution. Result rows are single int: 1 = collision, 0 = no collision.
	SqlCheckUniqueCollisions(
		schema *schema.EntitySchema, uniqueKeysToCheck [][]string, data schema.DynamicEntity,
	) (string, []any, error)
}

// Ensure interface implementation at compile time.
var _ QueryBuilder = (*PgQueryBuilder)(nil)

// PgQueryBuilder implements QueryBuilder for PostgreSQL.
type PgQueryBuilder struct {
}

func NewPgQueryBuilder() QueryBuilder {
	return &PgQueryBuilder{}
}

func (this *PgQueryBuilder) SqlCreateTable(schema *schema.EntitySchema) (string, error) {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(schema.Name())
	for _, col := range schema.Columns() {
		pgType, err := resolveGenericToPgType(col.ColumnType())
		if err != nil {
			return "", errors.Wrapf(err, "column '%s'", col.Name())
		}
		builder.Define(col.Name(), pgType, col.ColumnNullable())
	}

	if keys := schema.KeyColumns(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(keys), ", ")))
	}

	for _, unique := range schema.AllUniques() {
		if len(unique) == 0 {
			continue
		}
		builder.Define("UNIQUE", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(unique), ", ")))
	}

	sql, _ := builder.Build()
	return sql, nil
}

func (this *PgQueryBuilder) SqlSelectGraph(
	schema *schema.EntitySchema, graph schema.SearchGraph, columns []string,
) (string, error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 4)

	if len(columns) == 0 {
		mods = append(mods, sm.Columns(psql.Raw("*")))
	} else {
		columnExprs := make([]any, len(columns))
		for i, col := range columns {
			if _, ok := schema.Column(col); !ok {
				return "", errors.Errorf("unknown column '%s'", col)
			}
			columnExprs[i] = psql.Quote(col)
		}
		mods = append(mods, sm.Columns(columnExprs...))
	}

	mods = append(mods, sm.From(this.tableExpression(schema)))

	predicate, ok, err := this.graphExpression(
		schema, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
	if err != nil {
		return "", err
	}
	if ok {
		mods = append(mods, sm.Where(predicate))
	}

	orderMods, err := this.orderMods(schema, graph.GetOrder())
	if err != nil {
		return "", err
	}
	mods = append(mods, orderMods...)

	return this.buildSql(psql.Select(mods...))
}

func (this *PgQueryBuilder) tableExpression(schema *schema.EntitySchema) any {
	tableName := schema.TableName()
	if tableName == "" {
		tableName = schema.Name()
	}
	parts := strings.Split(tableName, ".")
	return psql.Quote(parts...)
}

func (this *PgQueryBuilder) SqlInsert(schema *schema.EntitySchema, data schema.DynamicEntity) (string, error) {
	return this.SqlInsertBulk(schema, []schema.DynamicEntity{data})
}

func (this *PgQueryBuilder) SqlInsertBulk(schema *schema.EntitySchema, rows []schema.DynamicEntity) (string, error) {
	prepared, err := this.rowsFrom(schema, rows, nil)
	if err != nil {
		return "", err
	}
	query, err := this.insertQuery(schema, prepared)
	if err != nil {
		return "", err
	}
	return this.buildSql(query)
}

func (this *PgQueryBuilder) SqlUpdateByPk(schema *schema.EntitySchema, data schema.DynamicEntity) (string, error) {
	if len(schema.PrimaryKeys()) == 0 {
		return "", errors.New("entity has no primary keys")
	}

	target, err := this.rowFromMap(schema, data, func(name string) bool {
		return !schema.IsPrimaryKey(name) && !schema.IsTenantKey(name)
	})
	if err != nil {
		return "", err
	}
	if len(target.columns) == 0 {
		return "", errors.New("no updatable columns provided")
	}

	lookup, err := this.rowForKeys(schema, data, schema.KeyColumns())
	if err != nil {
		return "", err
	}
	mods := this.buildUpdateMods(schema, target, lookup)
	return this.buildSql(psql.Update(mods...))
}

func (this *PgQueryBuilder) buildUpdateMods(
	schema *schema.EntitySchema, target rowData, lookup rowData,
) []bob.Mod[*dialect.UpdateQuery] {
	mods := []bob.Mod[*dialect.UpdateQuery]{um.Table(this.tableExpression(schema))}
	for i, col := range target.columns {
		mods = append(mods, um.SetCol(col).ToArg(target.values[i]))
	}
	for i, col := range lookup.columns {
		mods = append(mods, um.Where(psql.Quote(col).EQ(psql.Arg(lookup.values[i]))))
	}
	return mods
}

func (this *PgQueryBuilder) SqlDeleteEqual(schema *schema.EntitySchema, filters schema.DynamicEntity) (string, error) {
	if len(filters) == 0 {
		return "", errors.New("no filters provided")
	}

	row, err := this.rowFromMap(schema, filters, nil)
	if err != nil {
		return "", err
	}
	if len(row.columns) == 0 {
		return "", errors.New("no filters provided")
	}

	mods := []bob.Mod[*dialect.DeleteQuery]{dm.From(this.tableExpression(schema))}
	for i, col := range row.columns {
		mods = append(mods, dm.Where(psql.Quote(col).EQ(psql.Arg(row.values[i]))))
	}

	return this.buildSql(psql.Delete(mods...))
}

// SqlCheckUniqueCollisions builds SQL: SELECT 1 WHERE unique_key matches, else SELECT 0, UNION ALL per key.
// Returns (sql, args, nil). Each result row is 1 (collision) or 0 (no collision), in same order as uniqueKeysToCheck.
//
// Sample SQL with 3 unique keys (e.g. [email], [org_id, slug], [code]):
//
//	SELECT CASE WHEN EXISTS (SELECT 1 FROM "public"."users" WHERE "email" = $1) THEN 1 ELSE 0 END
//	UNION ALL
//	SELECT CASE WHEN EXISTS (SELECT 1 FROM "public"."users" WHERE "org_id" = $2 AND "slug" = $3) THEN 1 ELSE 0 END
//	UNION ALL
//	SELECT CASE WHEN EXISTS (SELECT 1 FROM "public"."users" WHERE "code" = $4) THEN 1 ELSE 0 END
func (this *PgQueryBuilder) SqlCheckUniqueCollisions(
	schema *schema.EntitySchema, uniqueKeysToCheck [][]string, data schema.DynamicEntity,
) (string, []any, error) {
	if len(uniqueKeysToCheck) == 0 {
		return "", nil, nil
	}

	tableRef := this.qualifiedTableRef(schema)
	tenantKey := schema.TenantKey()
	var args []any
	argIdx := 1
	var parts []string

	for _, uniqueFields := range uniqueKeysToCheck {
		if len(uniqueFields) == 0 {
			parts = append(parts, "SELECT 0")
			continue
		}

		columns := make([]string, 0, len(uniqueFields)+1)
		if tenantKey != "" {
			columns = append(columns, tenantKey)
		}
		columns = append(columns, uniqueFields...)

		values := make([]any, 0, len(columns))
		hasAll := true
		for _, col := range columns {
			v, ok := data[col]
			if !ok || v == nil {
				hasAll = false
				break
			}
			field, ok := schema.Column(col)
			if !ok {
				return "", nil, errors.Errorf("unknown column '%s'", col)
			}
			converted, err := this.convertValue(field, v)
			if err != nil {
				return "", nil, errors.Wrapf(err, "column '%s'", col)
			}
			values = append(values, converted)
		}
		if !hasAll {
			parts = append(parts, "SELECT 0")
			continue
		}

		conds := make([]string, len(columns))
		for i, col := range columns {
			conds[i] = fmt.Sprintf("%s = $%d", escapeIdentifier(col), argIdx)
			argIdx++
		}
		args = append(args, values...)
		whereClause := strings.Join(conds, " AND ")
		parts = append(parts,
			fmt.Sprintf("SELECT CASE WHEN EXISTS (SELECT 1 FROM %s WHERE %s) THEN 1 ELSE 0 END",
				tableRef, whereClause))
	}

	sql := strings.Join(parts, " UNION ALL ")
	return sql, args, nil
}

func (this *PgQueryBuilder) qualifiedTableRef(schema *schema.EntitySchema) string {
	tableName := schema.TableName()
	if tableName == "" {
		tableName = schema.Name()
	}
	parts := strings.Split(tableName, ".")
	escaped := make([]string, len(parts))
	for i, p := range parts {
		escaped[i] = escapeIdentifier(p)
	}
	return strings.Join(escaped, ".")
}

func escapeIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

type rowData struct {
	columns []string
	values  []any
}

func (this *PgQueryBuilder) rowsFrom(
	schema *schema.EntitySchema, rows []schema.DynamicEntity, filter func(string) bool,
) ([]rowData, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		item, err := this.rowFromMap(schema, row, filter)
		if err != nil {
			return nil, err
		}
		if len(item.columns) == 0 {
			return nil, errors.New("no columns provided")
		}
		if index == 0 {
			reference = item.columns
		} else if !slices.Equal(reference, item.columns) {
			return nil, errors.Errorf("row %d column mismatch", index)
		}
		prepared[index] = item
	}

	return prepared, nil
}

func (this *PgQueryBuilder) rowFromMap(
	schema *schema.EntitySchema, values schema.DynamicEntity, include func(string) bool,
) (rowData, error) {
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
		field, ok := schema.Column(key)
		if !ok {
			return rowData{}, errors.Errorf("unknown column '%s'", key)
		}
		converted, err := this.convertValue(field, values[key])
		if err != nil {
			return rowData{}, errors.Wrapf(err, "invalid value for column '%s'", key)
		}
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) rowForKeys(schema *schema.EntitySchema, values schema.DynamicEntity, keys []string) (rowData, error) {
	result := rowData{
		columns: make([]string, len(keys)),
		values:  make([]any, len(keys)),
	}

	for i, key := range keys {
		field, ok := schema.Column(key)
		if !ok {
			return rowData{}, errors.Errorf("unknown key column '%s'", key)
		}
		raw, ok := values[key]
		if !ok {
			return rowData{}, errors.Errorf("missing key '%s'", key)
		}
		converted, err := this.convertValue(field, raw)
		if err != nil {
			return rowData{}, errors.Wrapf(err, "invalid value for key '%s'", key)
		}
		result.columns[i] = key
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) insertQuery(schema *schema.EntitySchema, rows []rowData) (bob.Query, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}
	columns := rows[0].columns

	mods := []bob.Mod[*dialect.InsertQuery]{im.Into(this.tableExpression(schema), columns...)}
	for _, row := range rows {
		mods = append(mods, im.Values(psql.Arg(row.values...)))
	}

	return psql.Insert(mods...), nil
}

func (this *PgQueryBuilder) buildSql(query bob.Query) (string, error) {
	sql, args, err := bob.Build(context.Background(), query)
	if err != nil {
		return "", err
	}
	if len(args) == 0 {
		return sql, nil
	}

	iface := make([]interface{}, len(args))
	copy(iface, args)
	return sqlbuilder.PostgreSQL.Interpolate(sql, iface)
}

func (this *PgQueryBuilder) graphExpression(
	schema *schema.EntitySchema,
	condition *schema.Condition,
	and []schema.SearchNode,
	or []schema.SearchNode,
) (expr bob.Expression, ok bool, err error) {
	switch {
	case condition != nil:
		expr, err = this.conditionExpression(schema, *condition)
		return expr, err == nil, err
	case len(and) > 0:
		return this.combineNodes(schema, and, psql.And)
	case len(or) > 0:
		return this.combineNodes(schema, or, psql.Or)
	default:
		return nil, false, nil
	}
}

func (this *PgQueryBuilder) combineNodes(
	schema *schema.EntitySchema,
	nodes []schema.SearchNode,
	join func(...bob.Expression) psql.Expression,
) (expr bob.Expression, ok bool, err error) {
	expressions := make([]bob.Expression, 0, len(nodes))
	for _, node := range nodes {
		predicate, predicateOk, predicateErr := this.graphExpression(
			schema, node.GetCondition(), node.GetAnd(), node.GetOr())
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

func (this *PgQueryBuilder) conditionExpression(schema *schema.EntitySchema, cond schema.Condition) (bob.Expression, error) {
	field, expr, err := this.prepareCondition(schema, cond.Field())
	if err != nil {
		return nil, err
	}

	switch cond.Operator() {
	case schema.Equals, schema.NotEquals, schema.GreaterThan,
		schema.GreaterEqual, schema.LessThan, schema.LessEqual:
		return this.comparisonPredicate(expr, field, cond.Operator(), cond.Value())
	case schema.In, schema.NotIn:
		return this.collectionPredicate(expr, field, cond.Operator(), cond.Values())
	case schema.Contains, schema.NotContains, schema.StartsWith,
		schema.NotStartsWith, schema.EndsWith, schema.NotEndsWith:
		return this.stringPredicate(expr, field, cond.Operator(), cond.Value())
	case schema.IsSet, schema.IsNotSet:
		return this.nullPredicate(expr, cond.Operator()), nil
	default:
		return nil, errors.Errorf("unsupported operator '%s'", cond.Operator())
	}
}

func (this *PgQueryBuilder) prepareCondition(
	schema *schema.EntitySchema, fieldName string,
) (*schema.EntityField, psql.Expression, error) {
	if strings.Contains(fieldName, ".") {
		return nil, psql.Expression{}, errors.Errorf("nested fields not supported: %s", fieldName)
	}
	field, ok := schema.Column(fieldName)
	if !ok {
		return nil, psql.Expression{}, errors.Errorf("unknown column '%s'", fieldName)
	}
	return field, psql.Quote(field.Name()), nil
}

func (this *PgQueryBuilder) comparisonPredicate(
	expr psql.Expression, field *schema.EntityField, op schema.Operator, value any,
) (bob.Expression, error) {
	converted, err := this.convertValue(field, value)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted)

	switch op {
	case schema.Equals:
		return expr.EQ(arg), nil
	case schema.NotEquals:
		return expr.NE(arg), nil
	case schema.GreaterThan:
		return expr.GT(arg), nil
	case schema.GreaterEqual:
		return expr.GTE(arg), nil
	case schema.LessThan:
		return expr.LT(arg), nil
	case schema.LessEqual:
		return expr.LTE(arg), nil
	default:
		return nil, errors.Errorf("unsupported comparison operator '%s'", op)
	}
}

func (this *PgQueryBuilder) collectionPredicate(
	expr psql.Expression, field *schema.EntityField, op schema.Operator, values []any,
) (bob.Expression, error) {
	converted, err := this.convertValues(field, values)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted...)

	if op == schema.In {
		return expr.In(arg), nil
	}
	if op == schema.NotIn {
		return expr.NotIn(arg), nil
	}
	return nil, errors.Errorf("unsupported collection operator '%s'", op)
}

func (this *PgQueryBuilder) stringPredicate(
	expr psql.Expression, field *schema.EntityField, op schema.Operator, value any,
) (bob.Expression, error) {
	if columnCategoryFor(field.ColumnType()) != columnString {
		return nil, errors.Errorf("operator '%s' requires string column '%s'", op, field.Name())
	}

	converted, err := this.convertValue(field, value)
	if err != nil {
		return nil, err
	}
	pattern := stringPattern(fmt.Sprint(converted), op)
	if pattern == "" {
		return nil, errors.Errorf("unsupported string operator '%s'", op)
	}

	like := expr.ILike(psql.Arg(pattern))
	if strings.HasPrefix(string(op), "!") {
		return psql.Not(like), nil
	}
	return like, nil
}

func (this *PgQueryBuilder) nullPredicate(expr psql.Expression, op schema.Operator) bob.Expression {
	if op == schema.IsSet {
		return expr.IsNotNull()
	}
	return expr.IsNull()
}

func (this *PgQueryBuilder) convertValues(field *schema.EntityField, values []any) ([]any, error) {
	if len(values) == 0 {
		return nil, errors.Errorf("operator requires at least one value for column '%s'", field.Name())
	}
	converted := make([]any, len(values))
	for i, value := range values {
		next, err := this.convertValue(field, value)
		if err != nil {
			return nil, err
		}
		converted[i] = next
	}
	return converted, nil
}

func (this *PgQueryBuilder) orderMods(
	schema *schema.EntitySchema, order schema.SearchOrder,
) ([]bob.Mod[*dialect.SelectQuery], error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		fieldName := item[0]
		if strings.Contains(fieldName, ".") {
			return nil, errors.Errorf("nested order not supported: %s", fieldName)
		}
		if _, ok := schema.Column(fieldName); !ok {
			return nil, errors.Errorf("unknown order column '%s'", fieldName)
		}
		mod := sm.OrderBy(psql.Quote(fieldName))
		if item.Direction() == schema.Desc {
			mod = mod.Desc()
		} else {
			mod = mod.Asc()
		}
		mods = append(mods, mod)
	}
	return mods, nil
}

func (this *PgQueryBuilder) convertValue(field *schema.EntityField, value any) (any, error) {
	if value == nil {
		if field.IsNullable() {
			return nil, nil
		}
		return nil, errors.Errorf("column '%s' does not allow NULL", field.Name())
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if field.IsNullable() {
			return nil, nil
		}
		return nil, errors.Errorf("column '%s' does not allow NULL", field.Name())
	}
	if !valueAllowed(columnCategoryFor(field.ColumnType()), v) {
		return nil, errors.Errorf("column '%s': incompatible value type %T", field.Name(), v.Interface())
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

func resolveGenericToPgType(genericType string) (string, error) {
	switch genericType {
	case "email", "phone", "string", "secret", "url", "enumString", "ulid",
		"nikkiEtag", "nikkiLangCode", "nikkiModelId", "nikkiSlug":
		return "character varying", nil
	case "uuid":
		return "uuid", nil
	case "integer", "enumNumber":
		return "integer", nil
	case "float":
		return "double precision", nil
	case "boolean":
		return "boolean", nil
	case "date":
		return "date", nil
	case "time":
		return "time without time zone", nil
	case "dateTime":
		return "timestamptz", nil
	case "nikkiLangJson":
		return "jsonb", nil
	default:
		return "", errors.Errorf("unsupported generic type '%s'", genericType)
	}
}

func escapeIdentifiers(columns []string) []string {
	escaped := make([]string, len(columns))
	for i, col := range columns {
		escaped[i] = sqlbuilder.Escape(col)
	}
	return escaped
}

func stringPattern(value string, op schema.Operator) string {
	switch op {
	case schema.Contains, schema.NotContains:
		return "%" + value + "%"
	case schema.StartsWith, schema.NotStartsWith:
		return value + "%"
	case schema.EndsWith, schema.NotEndsWith:
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
	case strings.Contains(typ, "timestamp"), strings.Contains(typ, "timestamptz"),
		strings.Contains(typ, "date"), strings.Contains(typ, "time"):
		return columnTime
	case strings.Contains(typ, "char"), strings.Contains(typ, "text"), strings.Contains(typ, "uuid"),
		typ == "string" || typ == "email" || typ == "phone" || typ == "secret" || typ == "url" ||
			typ == "ulid" || typ == "enumstring" || strings.HasPrefix(typ, "nikki"):
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
		return isIntKind(v.Kind()) || isUintKind(v.Kind()) ||
			isFloatKind(v.Kind()) || isBigNumber(v.Interface())
	case columnTime:
		return v.Type() == timeType
	case columnJSON:
		return v.Kind() == reflect.Map
	default:
		return true
	}
}
