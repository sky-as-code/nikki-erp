package orm

import (
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
)

type QueryBuilder interface {
	SqlCreateTable(schema *schema.EntitySchema, registry *schema.EntityRegistry) (string, error)
	SqlSelectGraph(schema *schema.EntitySchema, graph schema.SearchGraph, columns []string) (string, error)
	SqlInsert(schema *schema.EntitySchema, data schema.DynamicFields) (string, error)
	SqlInsertBulk(schema *schema.EntitySchema, rows []schema.DynamicFields) (string, error)
	SqlUpdateByPk(schema *schema.EntitySchema, data schema.DynamicFields) (string, error)
	SqlDeleteEqual(schema *schema.EntitySchema, filters schema.DynamicFields) (string, error)
	// SqlCheckUniqueCollisions builds SQL that returns 1 per row where the unique key has a collision, else 0.
	// Input: uniqueKeysToCheck - subset of schema.AllUniques() where data has all values (no nil).
	// Returns (sql, args, nil) for execution. Result rows are single int: 1 = collision, 0 = no collision.
	SqlCheckUniqueCollisions(
		schema *schema.EntitySchema, uniqueKeysToCheck [][]string, data schema.DynamicFields,
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

func (this *PgQueryBuilder) SqlCreateTable(
	entSchema *schema.EntitySchema, registry *schema.EntityRegistry,
) (string, error) {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(pgQuote(entSchema.TableName()))
	if err := this.defineColumns(builder, entSchema); err != nil {
		return "", err
	}
	this.defineKeys(builder, entSchema)
	if err := this.defineForeignKeys(builder, entSchema, registry); err != nil {
		return "", err
	}
	sql, _ := builder.Build()
	return strings.TrimSuffix(sql, ";"), nil
}

func (this *PgQueryBuilder) defineColumns(
	builder *sqlbuilder.CreateTableBuilder, entSchema *schema.EntitySchema,
) error {
	for _, col := range entSchema.Columns() {
		pgType, err := resolveGenericToPgType(col.ColumnType())
		if err != nil {
			return errors.Wrapf(err, "column '%s'", col.Name())
		}
		builder.Define(col.Name(), pgType, col.ColumnNullable())
	}
	return nil
}

func (this *PgQueryBuilder) defineKeys(
	builder *sqlbuilder.CreateTableBuilder, entSchema *schema.EntitySchema,
) {
	if keys := entSchema.KeyColumns(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(keys), ", ")))
	}
	for _, unique := range entSchema.AllUniques() {
		if len(unique) == 0 {
			continue
		}
		name := pgQuote(fmt.Sprintf("%s_%s_ukey", entSchema.TableName(), strings.Join(unique, "_")))
		cols := fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(unique), ", "))
		builder.Define(fmt.Sprintf("CONSTRAINT %s UNIQUE", name), cols)
	}
}

func (this *PgQueryBuilder) defineForeignKeys(
	builder *sqlbuilder.CreateTableBuilder, entSchema *schema.EntitySchema, registry *schema.EntityRegistry,
) error {
	for _, rel := range entSchema.Relations() {
		if !isFkOwnerRelationType(rel.RelationType) {
			continue
		}
		destSchema := registry.Get(rel.DestEntityName)
		if destSchema == nil {
			return errors.Errorf("destination schema not found for foreign key on field '%s'", rel.SrcField)
		}
		fkName := pgQuote(fmt.Sprintf("%s_%s_fkey", entSchema.TableName(), rel.SrcField))
		fkBody := fmt.Sprintf("(%s) REFERENCES %s (%s) ON UPDATE %s ON DELETE %s",
			pgQuote(rel.SrcField), pgQuote(destSchema.TableName()), pgQuote(rel.DestField),
			rel.OnUpdate.Sql(), rel.OnDelete.Sql())
		builder.Define(fmt.Sprintf("CONSTRAINT %s FOREIGN KEY", fkName), fkBody)
	}
	return nil
}

func (this *PgQueryBuilder) SqlSelectGraph(
	entSchema *schema.EntitySchema, graph schema.SearchGraph, columns []string,
) (string, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	if len(columns) == 0 {
		sb.Select("*")
	} else {
		selectCols := make([]string, len(columns))
		for i, col := range columns {
			if _, ok := entSchema.Column(col); !ok {
				return "", errors.Errorf("unknown column '%s'", col)
			}
			selectCols[i] = pgQuote(col)
		}
		sb.Select(selectCols...)
	}

	sb.From(this.tableExpression(entSchema))

	predicate, ok, err := this.graphExpression(
		entSchema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
	if err != nil {
		return "", err
	}
	if ok {
		sb.Where(predicate)
	}

	orderExprs, err := this.orderExprs(entSchema, graph.GetOrder())
	if err != nil {
		return "", err
	}
	if len(orderExprs) > 0 {
		sb.OrderBy(orderExprs...)
	}

	sql, args := sb.Build()
	return interpolate(sql, args)
}

func (this *PgQueryBuilder) tableExpression(entSchema *schema.EntitySchema) string {
	tableName := entSchema.TableName()
	if tableName == "" {
		tableName = entSchema.Name()
	}
	return pgQuoteTable(strings.Split(tableName, ".")...)
}

func (this *PgQueryBuilder) SqlInsert(entSchema *schema.EntitySchema, data schema.DynamicFields) (string, error) {
	return this.SqlInsertBulk(entSchema, []schema.DynamicFields{data})
}

func (this *PgQueryBuilder) SqlInsertBulk(entSchema *schema.EntitySchema, rows []schema.DynamicFields) (string, error) {
	prepared, err := this.rowsFrom(entSchema, rows, nil)
	if err != nil {
		return "", err
	}
	return this.buildInsertSql(entSchema, prepared)
}

func (this *PgQueryBuilder) buildInsertSql(entSchema *schema.EntitySchema, rows []rowData) (string, error) {
	if len(rows) == 0 {
		return "", errors.New("no rows provided")
	}
	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
	ib.InsertInto(this.tableExpression(entSchema))
	ib.Cols(pgQuoteArr(rows[0].columns)...)
	for _, row := range rows {
		ib.Values(row.values...)
	}
	sql, args := ib.Build()
	return interpolate(sql, args)
}

func (this *PgQueryBuilder) SqlUpdateByPk(entSchema *schema.EntitySchema, data schema.DynamicFields) (string, error) {
	if len(entSchema.PrimaryKeys()) == 0 {
		return "", errors.New("entity has no primary keys")
	}

	target, err := this.rowFromMap(entSchema, data, func(name string) bool {
		return !entSchema.IsPrimaryKey(name) && !entSchema.IsTenantKey(name)
	})
	if err != nil {
		return "", err
	}
	if len(target.columns) == 0 {
		return "", errors.New("no updatable columns provided")
	}

	lookup, err := this.rowForKeys(entSchema, data, entSchema.KeyColumns())
	if err != nil {
		return "", err
	}
	return this.buildUpdateSql(entSchema, target, lookup)
}

func (this *PgQueryBuilder) buildUpdateSql(
	entSchema *schema.EntitySchema, target rowData, lookup rowData,
) (string, error) {
	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update(this.tableExpression(entSchema))
	assignments := make([]string, len(target.columns))
	for i, col := range target.columns {
		assignments[i] = ub.Assign(pgQuote(col), target.values[i])
	}
	ub.Set(assignments...)
	for i, col := range lookup.columns {
		ub.Where(ub.Equal(pgQuote(col), lookup.values[i]))
	}
	sql, args := ub.Build()
	return interpolate(sql, args)
}

func (this *PgQueryBuilder) SqlDeleteEqual(entSchema *schema.EntitySchema, filters schema.DynamicFields) (string, error) {
	if len(filters) == 0 {
		return "", errors.New("no filters provided")
	}

	row, err := this.rowFromMap(entSchema, filters, nil)
	if err != nil {
		return "", err
	}
	if len(row.columns) == 0 {
		return "", errors.New("no filters provided")
	}

	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	db.DeleteFrom(this.tableExpression(entSchema))
	for i, col := range row.columns {
		db.Where(db.Equal(pgQuote(col), row.values[i]))
	}
	sql, args := db.Build()
	return interpolate(sql, args)
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
	entSchema *schema.EntitySchema, uniqueKeysToCheck [][]string, data schema.DynamicFields,
) (string, []any, error) {
	if len(uniqueKeysToCheck) == 0 {
		return "", nil, nil
	}

	tableRef := this.tableExpression(entSchema)
	tenantKey := entSchema.TenantKey()
	var args []any
	argIdx := 1
	var parts []string

	for _, uniqueFields := range uniqueKeysToCheck {
		part, partArgs, err := this.buildUniqueCheckPart(entSchema, tableRef, tenantKey, uniqueFields, data, argIdx)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, part)
		args = append(args, partArgs...)
		argIdx += len(partArgs)
	}

	return strings.Join(parts, " UNION ALL "), args, nil
}

func (this *PgQueryBuilder) buildUniqueCheckPart(
	entSchema *schema.EntitySchema, tableRef string, tenantKey string,
	uniqueFields []string, data schema.DynamicFields, argIdx int,
) (string, []any, error) {
	if len(uniqueFields) == 0 {
		return "SELECT 0", nil, nil
	}
	columns := prependTenantKey(tenantKey, uniqueFields)
	values, hasAll, err := this.resolveColumnValues(entSchema, columns, data)
	if err != nil {
		return "", nil, err
	}
	if !hasAll {
		return "SELECT 0", nil, nil
	}
	return buildExistsCaseSql(tableRef, columns, argIdx), values, nil
}

func (this *PgQueryBuilder) resolveColumnValues(
	entSchema *schema.EntitySchema, columns []string, data schema.DynamicFields,
) ([]any, bool, error) {
	values := make([]any, 0, len(columns))
	for _, col := range columns {
		v, ok := data[col]
		if !ok || v == nil {
			return nil, false, nil
		}
		field, ok := entSchema.Column(col)
		if !ok {
			return nil, false, errors.Errorf("unknown column '%s'", col)
		}
		converted, err := this.convertValue(field, v)
		if err != nil {
			return nil, false, errors.Wrapf(err, "column '%s'", col)
		}
		values = append(values, converted)
	}
	return values, true, nil
}

type rowData struct {
	columns []string
	values  []any
}

func (this *PgQueryBuilder) rowsFrom(
	entSchema *schema.EntitySchema, rows []schema.DynamicFields, filter func(string) bool,
) ([]rowData, error) {
	if len(rows) == 0 {
		return nil, errors.New("no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		item, err := this.rowFromMap(entSchema, row, filter)
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
	entSchema *schema.EntitySchema, values schema.DynamicFields, include func(string) bool,
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
		field, ok := entSchema.Column(key)
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

func (this *PgQueryBuilder) rowForKeys(entSchema *schema.EntitySchema, values schema.DynamicFields, keys []string) (rowData, error) {
	result := rowData{
		columns: make([]string, len(keys)),
		values:  make([]any, len(keys)),
	}

	for i, key := range keys {
		field, ok := entSchema.Column(key)
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

func (this *PgQueryBuilder) graphExpression(
	entSchema *schema.EntitySchema,
	sb *sqlbuilder.SelectBuilder,
	condition *schema.Condition,
	and []schema.SearchNode,
	or []schema.SearchNode,
) (condStr string, ok bool, err error) {
	switch {
	case condition != nil:
		condStr, err = this.conditionExpression(entSchema, sb, *condition)
		return condStr, err == nil, err
	case len(and) > 0:
		return this.combineNodes(entSchema, sb, and, sb.And)
	case len(or) > 0:
		return this.combineNodes(entSchema, sb, or, sb.Or)
	default:
		return "", false, nil
	}
}

func (this *PgQueryBuilder) combineNodes(
	entSchema *schema.EntitySchema,
	sb *sqlbuilder.SelectBuilder,
	nodes []schema.SearchNode,
	join func(...string) string,
) (string, bool, error) {
	conditions := make([]string, 0, len(nodes))
	for _, node := range nodes {
		condStr, condOk, condErr := this.graphExpression(
			entSchema, sb, node.GetCondition(), node.GetAnd(), node.GetOr())
		if condErr != nil {
			return "", false, condErr
		}
		if condOk {
			conditions = append(conditions, condStr)
		}
	}
	if len(conditions) == 0 {
		return "", false, nil
	}
	return join(conditions...), true, nil
}

func (this *PgQueryBuilder) conditionExpression(
	entSchema *schema.EntitySchema,
	sb *sqlbuilder.SelectBuilder,
	cond schema.Condition,
) (string, error) {
	field, quotedField, err := this.prepareCondition(entSchema, cond.Field())
	if err != nil {
		return "", err
	}

	switch cond.Operator() {
	case schema.Equals, schema.NotEquals, schema.GreaterThan,
		schema.GreaterEqual, schema.LessThan, schema.LessEqual:
		return this.comparisonPredicate(sb, quotedField, field, cond.Operator(), cond.Value())
	case schema.In, schema.NotIn:
		return this.collectionPredicate(sb, quotedField, field, cond.Operator(), cond.Values())
	case schema.Contains, schema.NotContains, schema.StartsWith,
		schema.NotStartsWith, schema.EndsWith, schema.NotEndsWith:
		return this.stringPredicate(sb, quotedField, field, cond.Operator(), cond.Value())
	case schema.IsSet, schema.IsNotSet:
		return nullPredicate(sb, quotedField, cond.Operator()), nil
	default:
		return "", errors.Errorf("unsupported operator '%s'", cond.Operator())
	}
}

func (this *PgQueryBuilder) prepareCondition(
	entSchema *schema.EntitySchema, fieldName string,
) (*schema.EntityField, string, error) {
	if strings.Contains(fieldName, ".") {
		return nil, "", errors.Errorf("nested fields not supported: %s", fieldName)
	}
	field, ok := entSchema.Column(fieldName)
	if !ok {
		return nil, "", errors.Errorf("unknown column '%s'", fieldName)
	}
	return field, pgQuote(field.Name()), nil
}

func (this *PgQueryBuilder) comparisonPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *schema.EntityField, op schema.Operator, value any,
) (string, error) {
	converted, err := this.convertValue(field, value)
	if err != nil {
		return "", err
	}
	switch op {
	case schema.Equals:
		return sb.Equal(quotedField, converted), nil
	case schema.NotEquals:
		return sb.NotEqual(quotedField, converted), nil
	case schema.GreaterThan:
		return sb.GreaterThan(quotedField, converted), nil
	case schema.GreaterEqual:
		return sb.GreaterEqualThan(quotedField, converted), nil
	case schema.LessThan:
		return sb.LessThan(quotedField, converted), nil
	case schema.LessEqual:
		return sb.LessEqualThan(quotedField, converted), nil
	default:
		return "", errors.Errorf("unsupported comparison operator '%s'", op)
	}
}

func (this *PgQueryBuilder) collectionPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *schema.EntityField, op schema.Operator, values []any,
) (string, error) {
	converted, err := this.convertValues(field, values)
	if err != nil {
		return "", err
	}
	if op == schema.In {
		return sb.In(quotedField, converted...), nil
	}
	if op == schema.NotIn {
		return sb.NotIn(quotedField, converted...), nil
	}
	return "", errors.Errorf("unsupported collection operator '%s'", op)
}

func (this *PgQueryBuilder) stringPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *schema.EntityField, op schema.Operator, value any,
) (string, error) {
	if columnCategoryFor(field.ColumnType()) != columnString {
		return "", errors.Errorf("operator '%s' requires string column '%s'", op, field.Name())
	}
	converted, err := this.convertValue(field, value)
	if err != nil {
		return "", err
	}
	pattern := stringPattern(fmt.Sprint(converted), op)
	if pattern == "" {
		return "", errors.Errorf("unsupported string operator '%s'", op)
	}
	switch op {
	case schema.NotContains, schema.NotStartsWith, schema.NotEndsWith:
		return sb.NotILike(quotedField, pattern), nil
	default:
		return sb.ILike(quotedField, pattern), nil
	}
}

func nullPredicate(sb *sqlbuilder.SelectBuilder, quotedField string, op schema.Operator) string {
	if op == schema.IsSet {
		return sb.IsNotNull(quotedField)
	}
	return sb.IsNull(quotedField)
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

func (this *PgQueryBuilder) orderExprs(
	entSchema *schema.EntitySchema, order schema.SearchOrder,
) ([]string, error) {
	exprs := make([]string, 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		fieldName := item[0]
		if strings.Contains(fieldName, ".") {
			return nil, errors.Errorf("nested order not supported: %s", fieldName)
		}
		if _, ok := entSchema.Column(fieldName); !ok {
			return nil, errors.Errorf("unknown order column '%s'", fieldName)
		}
		dir := "ASC"
		if item.Direction() == schema.Desc {
			dir = "DESC"
		}
		exprs = append(exprs, fmt.Sprintf("%s %s", pgQuote(fieldName), dir))
	}
	return exprs, nil
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

func pgQuote(s string) string {
	return sqlbuilder.PostgreSQL.Quote(s)
}

func pgQuoteTable(parts ...string) string {
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = pgQuote(p)
	}
	return strings.Join(quoted, ".")
}

func pgQuoteArr(ss []string) []string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = pgQuote(s)
	}
	return quoted
}

func prependTenantKey(tenantKey string, fields []string) []string {
	if tenantKey == "" {
		return fields
	}
	cols := make([]string, 0, len(fields)+1)
	cols = append(cols, tenantKey)
	return append(cols, fields...)
}

func buildExistsCaseSql(tableRef string, columns []string, argIdx int) string {
	conds := make([]string, len(columns))
	for i, col := range columns {
		conds[i] = fmt.Sprintf("%s = $%d", pgQuote(col), argIdx+i)
	}
	whereClause := strings.Join(conds, " AND ")
	return fmt.Sprintf(
		"SELECT CASE WHEN EXISTS (SELECT 1 FROM %s WHERE %s) THEN 1 ELSE 0 END",
		tableRef, whereClause)
}

func interpolate(sql string, args []interface{}) (string, error) {
	if len(args) == 0 {
		return sql, nil
	}
	return sqlbuilder.PostgreSQL.Interpolate(sql, args)
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
