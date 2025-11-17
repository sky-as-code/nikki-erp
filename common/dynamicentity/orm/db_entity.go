package orm

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	eschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"go.bryk.io/pkg/errors"
)

type Column struct {
	Name     string
	Type     string
	Nullable string
}

var timeType = reflect.TypeOf(time.Time{})

type Operator string

const (
	Equals        Operator = "="
	NotEquals     Operator = "!="
	GreaterThan   Operator = ">"
	GreaterEqual  Operator = ">="
	LessThan      Operator = "<"
	LessEqual     Operator = "<="
	Contains      Operator = "*"
	NotContains   Operator = "!*"
	StartsWith    Operator = "^"
	NotStartsWith Operator = "!^"
	EndsWith      Operator = "$"
	NotEndsWith   Operator = "!$"
	In            Operator = "in"
	NotIn         Operator = "not_in"
	IsSet         Operator = "is_set"
	IsNotSet      Operator = "not_set"
)

type Condition []any

func NewCondition(field string, operator Operator, values ...any) Condition {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return Condition(arr)
}

func (c Condition) Field() string {
	return fmt.Sprint(c[0])
}

func (c Condition) Operator() Operator {
	if len(c) < 2 {
		return ""
	}
	if op, ok := c[1].(Operator); ok {
		return op
	}
	return Operator(fmt.Sprint(c[1]))
}

func (c Condition) Value() any {
	if len(c) < 3 {
		return nil
	}
	return c[2]
}

func (c Condition) Values() []any {
	if len(c) <= 2 {
		return nil
	}
	return c[2:]
}

type OrderDirection string

const (
	Asc  OrderDirection = "asc"
	Desc OrderDirection = "desc"
)

type SearchOrderItem []string

func (item SearchOrderItem) Direction() OrderDirection {
	if len(item) == 2 {
		return OrderDirection(strings.ToLower(item[1]))
	}
	return Asc
}

type SearchOrder []SearchOrderItem

type SearchGraph struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
	Order     SearchOrder  `json:"order,omitempty"`
}

type SearchNode struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
}

type DbEntity struct {
	Name        string
	Columns     []Column
	PrimaryKeys []string
	TenantKey   *string
	UniqueKeys  [][]string
}

type ResolveDbType func(field *eschema.EntityField) (string, error)

func NewPgDbEntity(schema *eschema.EntitySchema) (entity *DbEntity, err error) {
	return NewDbEntity(schema, resolvePostgresType)
}

func resolvePostgresType(field *eschema.EntityField) (string, error) {
	switch field.DataType() {
	case eschema.FieldDataTypeEmail,
		eschema.FieldDataTypePhone,
		eschema.FieldDataTypeString,
		eschema.FieldDataTypeSecret,
		eschema.FieldDataTypeUrl,
		eschema.FieldDataTypeEnumString:
		return "character varying", nil
	case eschema.FieldDataTypeUlid:
		return "character varying", nil
	case eschema.FieldDataTypeUuid:
		return "uuid", nil
	case eschema.FieldDataTypeInteger,
		eschema.FieldDataTypeEnumNumber:
		return "integer", nil
	case eschema.FieldDataTypeFloat:
		return "double precision", nil
	case eschema.FieldDataTypeBoolean:
		return "boolean", nil
	case eschema.FieldDataTypeDate:
		return "date", nil
	case eschema.FieldDataTypeTime:
		return "time without time zone", nil
	case eschema.FieldDataTypeDateTime:
		return "timestamptz", nil
	default:
		return "", fmt.Errorf("unsupported field data type '%s'", field.DataType())
	}
}

func NewDbEntity(schema *eschema.EntitySchema, resolveDbType ResolveDbType) (entity *DbEntity, err error) {
	name, err := requireName(schema.Name())
	if err != nil {
		return nil, err
	}

	columns, columnSet, primaryKeys, tenantKey, fieldUnique, err := buildColumns(schema.Fields(), name, resolveDbType)
	if err != nil {
		return nil, err
	}
	entityUnique, err := extractEntityUnique(schema.Rules(), columnSet)
	if err != nil {
		return nil, err
	}
	if len(primaryKeys) == 0 {
		return nil, fmt.Errorf("entity '%s' must define at least one primary key column", name)
	}

	entity = &DbEntity{
		Name:        name,
		Columns:     columns,
		PrimaryKeys: primaryKeys,
		UniqueKeys:  deduplicateKeySets(append(fieldUnique, entityUnique...)),
	}
	if tenantKey != "" {
		entity.TenantKey = &tenantKey
	}
	return entity, nil
}

func requireName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errors.New("entity name is required")
	}
	return name, nil
}

func validateFieldName(field *eschema.EntityField) error {
	columnName := strings.TrimSpace(field.Name())
	if columnName == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

func buildColumns(
	fields map[string]*eschema.EntityField,
	entityName string,
	resolveDbType ResolveDbType,
) (columns []Column, columnSet map[string]struct{}, primary []string, tenant string, uniques [][]string, err error) {
	columns = make([]Column, 0, len(fields))
	columnSet = make(map[string]struct{}, len(fields))
	uniques = make([][]string, 0)

	for _, field := range fields {
		if field == nil {
			continue
		}
		if err := validateFieldName(field); err != nil {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s': %w", entityName, err)
		}
		columnName := strings.TrimSpace(field.Name())
		if columnName == "" {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s' contains a field with empty name", entityName)
		}
		if _, exists := columnSet[columnName]; exists {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s' has duplicate column '%s'", entityName, columnName)
		}

		col, isUnique, isPrimary, isTenant, columnErr := makeColumn(field, columnName, entityName, resolveDbType)
		if columnErr != nil {
			return nil, nil, nil, "", nil, columnErr
		}

		columns = append(columns, *col)
		columnSet[columnName] = struct{}{}

		if isUnique {
			uniques = append(uniques, []string{columnName})
		}
		if isPrimary {
			primary = append(primary, columnName)
		}
		if isTenant {
			if tenant != "" && tenant != columnName {
				return nil, nil, nil, "", nil, errors.Errorf("%s must not be a tenant key because %s is already one", columnName, tenant)
			}
			tenant = columnName
		}
	}

	return columns, columnSet, primary, tenant, uniques, nil
}

func makeColumn(
	field *eschema.EntityField,
	columnName string,
	entityName string,
	resolveDbType ResolveDbType,
) (column *Column, unique bool, primary bool, tenant bool, err error) {
	columnType, err := resolveDbType(field)
	if err != nil {
		return nil, false, false, false, fmt.Errorf("entity '%s': %w", entityName, err)
	}

	nullable := "NULL"
	if field.IsRequired() {
		nullable = "NOT NULL"
	}

	for _, rule := range field.Rules() {
		ruleName := rule.RuleName()
		if ruleName == "" {
			continue
		}

		switch ruleName {
		case eschema.FieldRuleUniqueType:
			unique = true
		case eschema.FieldRulePrimaryType:
			primary = true
		case eschema.FieldRuleTenantType:
			tenant = true
		}
	}

	column = &Column{Name: columnName, Type: columnType, Nullable: nullable}
	return column, unique, primary, tenant, nil
}

func extractEntityUnique(rules []eschema.EntityRule, columnSet map[string]struct{}) (uniques [][]string, err error) {
	uniques = make([][]string, 0)
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		if fmt.Sprint(rule[0]) != string(eschema.EntityRuleNameUnique) {
			continue
		}
		groups, groupErr := collectRuleGroups(rule[1:], columnSet)
		if groupErr != nil {
			return nil, groupErr
		}
		uniques = append(uniques, groups...)
	}
	return uniques, nil
}

func collectRuleGroups(args []any, columnSet map[string]struct{}) ([][]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	groups := make([][]string, 0, len(args))
	for _, raw := range expandRuleArguments(args) {
		group := make([]string, 0, len(raw))
		for _, item := range raw {
			name := strings.TrimSpace(fmt.Sprint(item))
			if name == "" {
				continue
			}
			if _, ok := columnSet[name]; !ok {
				return nil, fmt.Errorf("unknown column reference '%s'", name)
			}
			group = append(group, name)
		}
		if len(group) == 0 {
			continue
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func expandRuleArguments(args []any) [][]any {
	result := make([][]any, 0, len(args))
	for _, arg := range args {
		switch v := arg.(type) {
		case []any:
			result = append(result, v)
		case []string:
			row := make([]any, len(v))
			for i, item := range v {
				row[i] = item
			}
			result = append(result, row)
		default:
			result = append(result, []any{arg})
		}
	}
	return result
}

func deduplicateKeySets(keys [][]string) [][]string {
	if len(keys) == 0 {
		return keys
	}

	seen := make(map[string]struct{}, len(keys))
	result := make([][]string, 0, len(keys))

	for _, key := range keys {
		if len(key) == 0 {
			continue
		}

		normalized := append([]string(nil), key...)
		sort.Strings(normalized)
		signature := strings.Join(normalized, "|")

		if _, exists := seen[signature]; exists {
			continue
		}

		seen[signature] = struct{}{}
		result = append(result, key)
	}

	return result
}

func escapeIdentifiers(columns []string) []string {
	escaped := make([]string, len(columns))
	for i, col := range columns {
		escaped[i] = sqlbuilder.Escape(col)
	}
	return escaped
}

type rowData struct {
	columns []string
	values  []any
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

func (e DbEntity) SqlCreateTable() string {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(e.Name)
	for _, col := range e.Columns {
		builder.Define(col.Name, col.Type, col.Nullable)
	}

	if keys := e.keyColumns(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(keys), ", ")))
	}

	for _, unique := range e.UniqueKeys {
		if len(unique) == 0 {
			continue
		}
		builder.Define("UNIQUE", fmt.Sprintf("(%s)", strings.Join(escapeIdentifiers(unique), ", ")))
	}

	sql, _ := builder.Build()
	return sql
}

func (e DbEntity) SqlSelectGraph(graph SearchGraph, columns []string) (string, error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, 4)

	if len(columns) == 0 {
		mods = append(mods, sm.Columns(psql.Raw("*")))
	} else {
		columnExprs := make([]any, len(columns))
		for i, col := range columns {
			if _, ok := e.column(col); !ok {
				return "", fmt.Errorf("unknown column '%s'", col)
			}
			columnExprs[i] = psql.Quote(col)
		}
		mods = append(mods, sm.Columns(columnExprs...))
	}

	mods = append(mods, sm.From(e.tableExpression()))

	predicate, ok, err := e.graphExpression(graph.Condition, graph.And, graph.Or)
	if err != nil {
		return "", err
	}
	if ok {
		mods = append(mods, sm.Where(predicate))
	}

	orderMods, err := e.orderMods(graph.Order)
	if err != nil {
		return "", err
	}
	mods = append(mods, orderMods...)

	return e.buildSQL(psql.Select(mods...))
}

func (e DbEntity) SqlInsertMap(data dmodel.EntityMap) (string, error) {
	return e.insertFromMaps([]dmodel.EntityMap{data})
}

func (e DbEntity) SqlInsertBulkMaps(rows []dmodel.EntityMap) (string, error) {
	return e.insertFromMaps(rows)
}

func (e DbEntity) SqlInsertStruct(payload any) (string, error) {
	m, err := e.structToMap(payload)
	if err != nil {
		return "", err
	}
	return e.SqlInsertMap(m)
}

func (e DbEntity) SqlInsertBulkStructs(items []any) (string, error) {
	maps, err := e.structSliceToMaps(items)
	if err != nil {
		return "", err
	}
	return e.SqlInsertBulkMaps(maps)
}

func (e DbEntity) SqlUpdateByPkMap(data dmodel.EntityMap) (string, error) {
	return e.updateFromMap(data)
}

func (e DbEntity) SqlUpdateByPkStruct(payload any) (string, error) {
	m, err := e.structToMap(payload)
	if err != nil {
		return "", err
	}
	return e.updateFromMap(m)
}

func (e DbEntity) SqlDeleteEqualStruct(filters dmodel.EntityMap) (string, error) {
	if len(filters) == 0 {
		return "", errors.New("no filters provided")
	}
	if err := e.ensureTenantKeyInMap(filters); err != nil {
		return "", err
	}

	row, err := e.rowFromMap(filters, nil)
	if err != nil {
		return "", err
	}
	if len(row.columns) == 0 {
		return "", errors.New("no filters provided")
	}

	mods := []bob.Mod[*dialect.DeleteQuery]{dm.From(e.tableExpression())}
	for i, col := range row.columns {
		mods = append(mods, dm.Where(psql.Quote(col).EQ(psql.Arg(row.values[i]))))
	}

	return e.buildSQL(psql.Delete(mods...))
}

func (e DbEntity) insertFromMaps(rows []dmodel.EntityMap) (string, error) {
	prepared, err := e.rowsFromMaps(rows, nil)
	if err != nil {
		return "", err
	}
	query, err := e.insertQuery(prepared)
	if err != nil {
		return "", err
	}
	return e.buildSQL(query)
}

func (e DbEntity) updateFromMap(data dmodel.EntityMap) (string, error) {
	if len(e.PrimaryKeys) == 0 {
		return "", errors.New("entity has no primary keys")
	}
	if err := e.ensureTenantKeyInMap(data); err != nil {
		return "", err
	}

	target, err := e.rowFromMap(data, func(name string) bool {
		return !e.isPrimaryKey(name) && !e.isTenantKey(name)
	})
	if err != nil {
		return "", err
	}
	if len(target.columns) == 0 {
		return "", errors.New("no updatable columns provided")
	}

	lookup, err := e.rowForKeys(data, e.keyColumns())
	if err != nil {
		return "", err
	}
	mods := []bob.Mod[*dialect.UpdateQuery]{um.Table(e.tableExpression())}
	for i, col := range target.columns {
		mods = append(mods, um.SetCol(col).ToArg(target.values[i]))
	}
	for i, col := range lookup.columns {
		mods = append(mods, um.Where(psql.Quote(col).EQ(psql.Arg(lookup.values[i]))))
	}

	return e.buildSQL(psql.Update(mods...))
}

func (e DbEntity) rowsFromMaps(rows []dmodel.EntityMap, filter func(string) bool) ([]rowData, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		if err := e.ensureTenantKeyInMap(row); err != nil {
			return nil, err
		}
		item, err := e.rowFromMap(row, filter)
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

func (e DbEntity) rowFromMap(values dmodel.EntityMap, include func(string) bool) (rowData, error) {
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
		col, ok := e.column(key)
		if !ok {
			return rowData{}, fmt.Errorf("unknown column '%s'", key)
		}
		converted, err := e.convertValue(col, values[key])
		if err != nil {
			return rowData{}, fmt.Errorf("invalid value for column '%s': %w", key, err)
		}
		result.values[i] = converted
	}

	return result, nil
}

func (e DbEntity) rowForKeys(values dmodel.EntityMap, keys []string) (rowData, error) {
	result := rowData{
		columns: make([]string, len(keys)),
		values:  make([]any, len(keys)),
	}

	for i, key := range keys {
		col, ok := e.column(key)
		if !ok {
			return rowData{}, fmt.Errorf("unknown key column '%s'", key)
		}
		raw, ok := values[key]
		if !ok {
			return rowData{}, fmt.Errorf("missing key '%s'", key)
		}
		converted, err := e.convertValue(col, raw)
		if err != nil {
			return rowData{}, fmt.Errorf("invalid value for key '%s': %w", key, err)
		}
		result.columns[i] = key
		result.values[i] = converted
	}

	return result, nil
}

func (e DbEntity) insertQuery(rows []rowData) (bob.Query, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}
	columns := rows[0].columns

	mods := []bob.Mod[*dialect.InsertQuery]{im.Into(e.tableExpression(), columns...)}
	for _, row := range rows {
		mods = append(mods, im.Values(psql.Arg(row.values...)))
	}

	return psql.Insert(mods...), nil
}

func (e DbEntity) structToMap(payload any) (dmodel.EntityMap, error) {
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
	return exportStructFields(value)
}

func exportStructFields(value reflect.Value) (dmodel.EntityMap, error) {
	result := make(dmodel.EntityMap)
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag := strings.Split(field.Tag.Get(eschema.SchemaStructTag), ",")[0]
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

func (e DbEntity) structSliceToMaps(items []any) ([]dmodel.EntityMap, error) {
	if len(items) == 0 {
		return nil, errors.New("no payload provided")
	}
	result := make([]dmodel.EntityMap, len(items))
	for i, item := range items {
		m, err := e.structToMap(item)
		if err != nil {
			return nil, err
		}
		result[i] = m
	}
	return result, nil
}

func (e DbEntity) buildSQL(query bob.Query) (string, error) {
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

func (e DbEntity) graphExpression(
	condition *Condition,
	and []SearchNode,
	or []SearchNode,
) (expr bob.Expression, ok bool, err error) {
	switch {
	case condition != nil:
		expr, err = e.conditionExpression(*condition)
		return expr, err == nil, err
	case len(and) > 0:
		return e.combineNodes(and, psql.And)
	case len(or) > 0:
		return e.combineNodes(or, psql.Or)
	default:
		return nil, false, nil
	}
}

func (e DbEntity) combineNodes(
	nodes []SearchNode,
	join func(...bob.Expression) psql.Expression,
) (expr bob.Expression, ok bool, err error) {
	expressions := make([]bob.Expression, 0, len(nodes))
	for _, node := range nodes {
		predicate, predicateOk, predicateErr := e.graphExpression(node.Condition, node.And, node.Or)
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

func (e DbEntity) conditionExpression(cond Condition) (bob.Expression, error) {
	col, expr, err := e.prepareCondition(cond.Field())
	if err != nil {
		return nil, err
	}

	switch cond.Operator() {
	case Equals, NotEquals, GreaterThan, GreaterEqual, LessThan, LessEqual:
		return e.comparisonPredicate(expr, col, cond.Operator(), cond.Value())
	case In, NotIn:
		return e.collectionPredicate(expr, col, cond.Operator(), cond.Values())
	case Contains, NotContains, StartsWith, NotStartsWith, EndsWith, NotEndsWith:
		return e.stringPredicate(expr, col, cond.Operator(), cond.Value())
	case IsSet, IsNotSet:
		return e.nullPredicate(expr, cond.Operator()), nil
	default:
		return nil, fmt.Errorf("unsupported operator '%s'", cond.Operator())
	}
}

func (e DbEntity) prepareCondition(field string) (Column, psql.Expression, error) {
	if strings.Contains(field, ".") {
		return Column{}, psql.Expression{}, fmt.Errorf("nested fields not supported: %s", field)
	}
	col, ok := e.column(field)
	if !ok {
		return Column{}, psql.Expression{}, fmt.Errorf("unknown column '%s'", field)
	}
	return col, psql.Quote(col.Name), nil
}

func (e DbEntity) comparisonPredicate(expr psql.Expression, col Column, op Operator, value any) (bob.Expression, error) {
	converted, err := e.convertValue(col, value)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted)

	switch op {
	case Equals:
		return expr.EQ(arg), nil
	case NotEquals:
		return expr.NE(arg), nil
	case GreaterThan:
		return expr.GT(arg), nil
	case GreaterEqual:
		return expr.GTE(arg), nil
	case LessThan:
		return expr.LT(arg), nil
	case LessEqual:
		return expr.LTE(arg), nil
	default:
		return nil, fmt.Errorf("unsupported comparison operator '%s'", op)
	}
}

func (e DbEntity) collectionPredicate(expr psql.Expression, col Column, op Operator, values []any) (bob.Expression, error) {
	converted, err := e.convertValues(col, values)
	if err != nil {
		return nil, err
	}
	arg := psql.Arg(converted...)

	if op == In {
		return expr.In(arg), nil
	}
	if op == NotIn {
		return expr.NotIn(arg), nil
	}
	return nil, fmt.Errorf("unsupported collection operator '%s'", op)
}

func (e DbEntity) stringPredicate(expr psql.Expression, col Column, op Operator, value any) (bob.Expression, error) {
	if columnCategoryFor(col.Type) != columnString {
		return nil, fmt.Errorf("operator '%s' requires string column '%s'", op, col.Name)
	}

	converted, err := e.convertValue(col, value)
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

func (e DbEntity) nullPredicate(expr psql.Expression, op Operator) bob.Expression {
	if op == IsSet {
		return expr.IsNotNull()
	}
	return expr.IsNull()
}

func stringPattern(value string, op Operator) string {
	switch op {
	case Contains, NotContains:
		return "%" + value + "%"
	case StartsWith, NotStartsWith:
		return value + "%"
	case EndsWith, NotEndsWith:
		return "%" + value
	default:
		return ""
	}
}

func (e DbEntity) convertValues(col Column, values []any) ([]any, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("operator requires at least one value for column '%s'", col.Name)
	}
	converted := make([]any, len(values))
	for i, value := range values {
		next, err := e.convertValue(col, value)
		if err != nil {
			return nil, err
		}
		converted[i] = next
	}
	return converted, nil
}

func (e DbEntity) orderMods(order SearchOrder) ([]bob.Mod[*dialect.SelectQuery], error) {
	mods := make([]bob.Mod[*dialect.SelectQuery], 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		field := item[0]
		if strings.Contains(field, ".") {
			return nil, fmt.Errorf("nested order not supported: %s", field)
		}
		if _, ok := e.column(field); !ok {
			return nil, fmt.Errorf("unknown order column '%s'", field)
		}
		mod := sm.OrderBy(psql.Quote(field))
		if item.Direction() == Desc {
			mod = mod.Desc()
		} else {
			mod = mod.Asc()
		}
		mods = append(mods, mod)
	}
	return mods, nil
}

func (e DbEntity) ensureTenantKeyInMap(values dmodel.EntityMap) error {
	key := e.tenantKey()
	if key == "" {
		return nil
	}
	if _, ok := values[key]; !ok {
		return fmt.Errorf("missing tenant key '%s'", key)
	}
	return nil
}

func (e DbEntity) tenantKey() string {
	if e.TenantKey == nil {
		return ""
	}
	return *e.TenantKey
}

func (e DbEntity) column(name string) (Column, bool) {
	for _, col := range e.Columns {
		if col.Name == name {
			return col, true
		}
	}
	return Column{}, false
}

func (e DbEntity) keyColumns() []string {
	keys := append([]string{}, e.PrimaryKeys...)
	if tk := e.tenantKey(); tk != "" && !contains(keys, tk) {
		keys = append(keys, tk)
	}
	return keys
}

func (e DbEntity) isPrimaryKey(name string) bool {
	return contains(e.PrimaryKeys, name)
}

func (e DbEntity) isTenantKey(name string) bool {
	return e.tenantKey() == name
}

func (e DbEntity) tableExpression() any {
	parts := strings.Split(e.Name, ".")
	return psql.Quote(parts...)
}

func (c Column) isNullable() bool {
	return !strings.EqualFold(c.Nullable, "NOT NULL")
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

func (e DbEntity) convertValue(col Column, value any) (any, error) {
	if value == nil {
		if col.isNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("column '%s' does not allow NULL", col.Name)
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if col.isNullable() {
			return nil, nil
		}
		return nil, fmt.Errorf("column '%s' does not allow NULL", col.Name)
	}
	if !valueAllowed(columnCategoryFor(col.Type), v) {
		return nil, fmt.Errorf("column '%s': incompatible value type %T", col.Name, v.Interface())
	}
	return v.Interface(), nil
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

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
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
