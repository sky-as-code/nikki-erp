package orm

import (
	stdErr "errors"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
)

// SqlSelectGraphOpts holds optional parameters for SqlSelectGraph.
type SqlSelectGraphOpts struct {
	// Columns limits which columns to fetch; empty means all columns (*).
	Columns []string
	// Page is the 0-based page index used to compute the OFFSET (OFFSET = Page * Size).
	// Ignored when Size is 0.
	Page int
	// Size sets the LIMIT clause. 0 means no limit/offset is applied.
	Size int
}

// SqlCheckUniqueCollisionsData holds parameterized SQL and arguments from SqlCheckUniqueCollisions.
type SqlCheckUniqueCollisionsData struct {
	Sql  string
	Args []any
}

type QueryBuilder interface {
	SqlCreateTable(schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry) (*crud.OpResult[string], error)
	SqlSelectGraph(schema *dmodel.ModelSchema, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts) (
		*crud.OpResult[string], error)
	// SqlCountGraph builds SELECT COUNT(*) with the same WHERE as SqlSelectGraph (no ORDER BY, LIMIT, OFFSET).
	SqlCountGraph(schema *dmodel.ModelSchema, graph *dmodel.SearchGraph) (*crud.OpResult[string], error)
	SqlInsert(schema *dmodel.ModelSchema, data dmodel.DynamicFields) (*crud.OpResult[string], error)
	SqlInsertBulk(schema *dmodel.ModelSchema, rows []dmodel.DynamicFields) (*crud.OpResult[string], error)
	SqlUpdateEqual(schema *dmodel.ModelSchema, data dmodel.DynamicFields, filters dmodel.DynamicFields) (
		*crud.OpResult[string], error)
	// SqlDeleteEqual generates a DELETE statement with the given filters using only equal predicates.
	// This DELETE statement can result in one or multiple rows being deleted.
	SqlDeleteEqual(schema *dmodel.ModelSchema, filters dmodel.DynamicFields) (*crud.OpResult[string], error)
	// SqlCheckUniqueCollisions builds SQL that returns 1 per row where the unique key has a collision, else 0.
	// Input: uniqueKeysToCheck - subset of dmodel.AllUniques() where data has all values (no nil).
	// Data.Sql / Data.Args are for execution. Result rows are single int: 1 = collision, 0 = no collision.
	SqlCheckUniqueCollisions(
		schema *dmodel.ModelSchema, uniqueKeysToCheck [][]string, data dmodel.DynamicFields,
	) (*crud.OpResult[SqlCheckUniqueCollisionsData], error)
	// RegisterPredefinedPredicate binds a filter field name to a treatment. Omit schemaName to apply to all schemas
	// (PredefinedPredicateAllSchemas). Operator uses dmodel.Operator values from model/database condition operators.
	RegisterPredefinedPredicate(fieldName string, treatment PredefinedPredicateTreatment, schemaName ...string) error
	GetPredefinedPredicate(fieldName string, schemaName string) PredefinedPredicateTreatment
}

// Ensure interface implementation at compile time.
var _ QueryBuilder = (*PgQueryBuilder)(nil)

// PgQueryBuilder implements QueryBuilder for PostgreSQL.
type PgQueryBuilder struct {
	predefinedPredicates map[string]map[string]PredefinedPredicateTreatment
}

func NewPgQueryBuilder() QueryBuilder {
	return &PgQueryBuilder{
		predefinedPredicates: make(map[string]map[string]PredefinedPredicateTreatment),
	}
}

func (this *PgQueryBuilder) SqlCreateTable(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) (*crud.OpResult[string], error) {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(pgQuote(schema.TableName()))
	if err := this.defineColumns(builder, schema); err != nil {
		return nil, err
	}
	this.defineKeys(builder, schema)
	if err := this.defineForeignKeys(builder, schema, registry); err != nil {
		return nil, err
	}
	sql, _ := builder.Build()
	out := strings.TrimSuffix(sql, ";")
	return &crud.OpResult[string]{Data: out, IsEmpty: false}, nil
}

func (this *PgQueryBuilder) defineColumns(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema,
) error {
	for _, col := range schema.Columns() {
		pgType, err := resolveGenericToPgType(col.ColumnType())
		if err != nil {
			return errors.Wrapf(err, "defineColumns: column '%s'", col.Name())
		}
		builder.Define(col.Name(), pgType, col.ColumnNullable())
	}
	return nil
}

func (this *PgQueryBuilder) defineKeys(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema,
) {
	if keys := schema.KeyColumns(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(keys), ", ")))
	}
	for _, unique := range schema.AllUniques() {
		if len(unique) == 0 {
			continue
		}
		name := pgQuote(fmt.Sprintf("%s_%s_ukey", schema.TableName(), strings.Join(unique, "_")))
		cols := fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(unique), ", "))
		builder.Define(fmt.Sprintf("CONSTRAINT %s UNIQUE", name), cols)
	}
}

func (this *PgQueryBuilder) defineForeignKeys(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) error {
	for _, rel := range schema.Relations() {
		if !isFkOwnerRelationType(rel.RelationType) {
			continue
		}
		destSchema := registry.Get(rel.DestSchemaName)
		if destSchema == nil {
			return errors.Errorf(
				"defineForeignKeys: destination schema not found for foreign key on field '%s'", rel.SrcField)
		}
		fkName := pgQuote(fmt.Sprintf("%s_%s_fkey", schema.TableName(), rel.SrcField))
		fkBody := fmt.Sprintf("(%s) REFERENCES %s (%s) ON UPDATE %s ON DELETE %s",
			pgQuote(rel.SrcField), pgQuote(destSchema.TableName()), pgQuote(rel.DestField),
			rel.OnUpdate.Sql(), rel.OnDelete.Sql())
		builder.Define(fmt.Sprintf("CONSTRAINT %s FOREIGN KEY", fkName), fkBody)
	}
	return nil
}

func (this *PgQueryBuilder) SqlSelectGraph(
	schema *dmodel.ModelSchema, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (*crud.OpResult[string], error) {
	sql, err := this.buildSqlSelectGraph(schema, graph, opts)
	if err != nil {
		return extractClientErr[string](err)
	}
	return &crud.OpResult[string]{Data: sql, IsEmpty: false}, nil
}

func (this *PgQueryBuilder) buildSqlSelectGraph(
	schema *dmodel.ModelSchema, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (string, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	if err := this.applySelectColumns(sb, schema, opts.Columns); err != nil {
		return "", err
	}
	sb.From(this.tableExpression(schema))
	if graph != nil {
		predicate, ok, err := this.graphExpression(
			schema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
		if err != nil {
			return "", err
		}
		if ok {
			sb.Where(predicate)
		}
		orderExprs, err := this.orderExprs(schema, graph.GetOrder())
		if err != nil {
			return "", err
		}
		if len(orderExprs) > 0 {
			sb.OrderBy(orderExprs...)
		}
	}
	this.applyPagination(sb, opts.Page, opts.Size)
	sql, args := sb.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", errors.Wrap(ierr, "buildSqlSelectGraph: interpolate")
	}
	return out, nil
}

func (this *PgQueryBuilder) SqlCountGraph(
	schema *dmodel.ModelSchema, graph *dmodel.SearchGraph,
) (*crud.OpResult[string], error) {
	sql, err := this.buildSqlCountGraph(schema, graph)
	if err != nil {
		return extractClientErr[string](err)
	}
	return &crud.OpResult[string]{Data: sql, IsEmpty: false}, nil
}

func (this *PgQueryBuilder) buildSqlCountGraph(
	schema *dmodel.ModelSchema, graph *dmodel.SearchGraph,
) (string, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("COUNT(*)")
	sb.From(this.tableExpression(schema))
	if graph != nil {
		predicate, ok, err := this.graphExpression(
			schema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
		if err != nil {
			return "", err
		}
		if ok {
			sb.Where(predicate)
		}
	}
	sql, args := sb.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", errors.Wrap(ierr, "buildSqlCountGraph: interpolate")
	}
	return out, nil
}

func (this *PgQueryBuilder) applySelectColumns(
	sb *sqlbuilder.SelectBuilder, schema *dmodel.ModelSchema, columns []string,
) error {
	if len(columns) == 0 {
		sb.Select("*")
		return nil
	}
	selectCols := make([]string, len(columns))
	for i, col := range columns {
		field, ok := schema.Column(col)
		if !ok || field.IsVirtualModelField() {
			return errors.Wrap(&errClientUnknownField{Field: col}, "applySelectColumns")
		}
		selectCols[i] = pgQuote(col)
	}
	sb.Select(selectCols...)
	return nil
}

func (this *PgQueryBuilder) applyPagination(sb *sqlbuilder.SelectBuilder, page, size int) {
	if size <= 0 {
		return
	}
	sb.Limit(size)
	if page > 0 {
		sb.Offset(page * size)
	}
}

func (this *PgQueryBuilder) tableExpression(schema *dmodel.ModelSchema) string {
	tableName := schema.TableName()
	if tableName == "" {
		tableName = schema.Name()
	}
	return pgQuoteTable(strings.Split(tableName, ".")...)
}

func (this *PgQueryBuilder) SqlInsert(schema *dmodel.ModelSchema, data dmodel.DynamicFields) (
	*crud.OpResult[string], error,
) {
	return this.SqlInsertBulk(schema, []dmodel.DynamicFields{data})
}

func (this *PgQueryBuilder) SqlInsertBulk(schema *dmodel.ModelSchema, rows []dmodel.DynamicFields) (
	*crud.OpResult[string], error,
) {
	prepared, err := this.rowsFrom(schema, rows, nil)
	if err != nil {
		return extractClientErr[string](err)
	}
	sql, ierr := this.buildInsertSql(schema, prepared)
	if ierr != nil {
		return extractClientErr[string](ierr)
	}
	return &crud.OpResult[string]{Data: sql, IsEmpty: false}, nil
}

func (this *PgQueryBuilder) buildInsertSql(schema *dmodel.ModelSchema, rows []rowData) (string, error) {
	if len(rows) == 0 {
		return "", errors.New("buildInsertSql: no rows provided")
	}
	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
	ib.InsertInto(this.tableExpression(schema))
	ib.Cols(pgQuoteArr(rows[0].columns)...)
	for _, row := range rows {
		ib.Values(row.values...)
	}
	sql, args := ib.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", errors.Wrap(ierr, "buildInsertSql: interpolate")
	}
	return out, nil
}

func (this *PgQueryBuilder) SqlUpdateEqual(
	schema *dmodel.ModelSchema,
	data dmodel.DynamicFields,
	filters dmodel.DynamicFields,
) (*crud.OpResult[string], error) {
	if len(filters) == 0 {
		return nil, errors.New("SqlUpdateEqual: no filters provided")
	}

	target, err := this.rowFromMap(schema, data, func(name string) bool {
		return !schema.IsPrimaryKey(name) && !schema.IsTenantKey(name)
	})
	if err != nil {
		return extractClientErr[string](err)
	}
	if len(target.columns) == 0 {
		return nil, errors.New("SqlUpdateEqual: no updatable columns provided")
	}

	lookup, err := this.rowFromMap(schema, filters, nil)
	if err != nil {
		return extractClientErr[string](err)
	}
	sql, ierr := this.buildUpdateSql(schema, target, lookup)
	if ierr != nil {
		return extractClientErr[string](ierr)
	}
	return &crud.OpResult[string]{Data: sql, IsEmpty: false}, nil
}

func (this *PgQueryBuilder) buildUpdateSql(
	schema *dmodel.ModelSchema, target rowData, lookup rowData,
) (string, error) {
	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update(this.tableExpression(schema))
	assignments := make([]string, len(target.columns))
	for i, col := range target.columns {
		assignments[i] = ub.Assign(pgQuote(col), target.values[i])
	}
	ub.Set(assignments...)
	for i, col := range lookup.columns {
		if lookup.values[i] == nil {
			ub.Where(ub.IsNull(pgQuote(col)))
		} else {
			ub.Where(ub.Equal(pgQuote(col), lookup.values[i]))
		}
	}
	sql, args := ub.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", errors.Wrap(ierr, "buildUpdateSql: interpolate")
	}
	return out, nil
}

func (this *PgQueryBuilder) SqlDeleteEqual(schema *dmodel.ModelSchema, filters dmodel.DynamicFields) (
	*crud.OpResult[string], error,
) {
	if len(filters) == 0 {
		return nil, errors.New("SqlDeleteEqual: no filters provided")
	}

	row, err := this.rowFromMap(schema, filters, nil)
	if err != nil {
		return extractClientErr[string](err)
	}
	if len(row.columns) == 0 {
		return nil, errors.New("SqlDeleteEqual: no filters provided")
	}

	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	db.DeleteFrom(this.tableExpression(schema))
	for i, col := range row.columns {
		db.Where(db.Equal(pgQuote(col), row.values[i]))
	}
	sql, args := db.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return extractClientErr[string](errors.Wrap(ierr, "SqlDeleteEqual: interpolate"))
	}
	return &crud.OpResult[string]{Data: out, IsEmpty: false}, nil
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
	schema *dmodel.ModelSchema, uniqueKeysToCheck [][]string, data dmodel.DynamicFields,
) (*crud.OpResult[SqlCheckUniqueCollisionsData], error) {
	if len(uniqueKeysToCheck) == 0 {
		return &crud.OpResult[SqlCheckUniqueCollisionsData]{IsEmpty: true}, nil
	}

	tableRef := this.tableExpression(schema)
	tenantKey := schema.TenantKey()
	var args []any
	argIdx := 1
	var parts []string

	for _, uniqueFields := range uniqueKeysToCheck {
		part, partArgs, err := this.buildUniqueCheckPart(schema, tableRef, tenantKey, uniqueFields, data, argIdx)
		if err != nil {
			return extractClientErr[SqlCheckUniqueCollisionsData](err)
		}
		parts = append(parts, part)
		args = append(args, partArgs...)
		argIdx += len(partArgs)
	}

	joined := strings.Join(parts, " UNION ALL ")
	return &crud.OpResult[SqlCheckUniqueCollisionsData]{
		Data:    SqlCheckUniqueCollisionsData{Sql: joined, Args: args},
		IsEmpty: false,
	}, nil
}

func (this *PgQueryBuilder) buildUniqueCheckPart(
	schema *dmodel.ModelSchema, tableRef string, tenantKey string,
	uniqueFields []string, data dmodel.DynamicFields, argIdx int,
) (string, []any, error) {
	if len(uniqueFields) == 0 {
		return "SELECT 0", nil, nil
	}
	columns := prependTenantKey(tenantKey, uniqueFields)
	values, hasAll, err := this.resolveColumnValues(schema, columns, data)
	if err != nil {
		return "", nil, err
	}
	if !hasAll {
		return "SELECT 0", nil, nil
	}
	return buildExistsCaseSql(tableRef, columns, argIdx), values, nil
}

func (this *PgQueryBuilder) resolveColumnValues(
	schema *dmodel.ModelSchema, columns []string, data dmodel.DynamicFields,
) ([]any, bool, error) {
	values := make([]any, 0, len(columns))
	for _, col := range columns {
		v, ok := data[col]
		if !ok || v == nil {
			return nil, false, nil
		}
		field, ok := schema.Column(col)
		if !ok || field.IsVirtualModelField() {
			return nil, false, errors.Wrap(&errClientUnknownField{Field: col}, "resolveColumnValues")
		}
		converted, err := this.convertValue(field, v)
		if err != nil {
			return nil, false, errors.Wrapf(err, "resolveColumnValues: column '%s'", col)
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
	schema *dmodel.ModelSchema, rows []dmodel.DynamicFields, filter func(string) bool,
) ([]rowData, error) {
	if len(rows) == 0 {
		return nil, errors.New("rowsFrom: no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		item, err := this.rowFromMap(schema, row, filter)
		if err != nil {
			return nil, err
		}
		if len(item.columns) == 0 {
			return nil, errors.New("rowsFrom: no columns provided")
		}
		if index == 0 {
			reference = item.columns
		} else if !slices.Equal(reference, item.columns) {
			return nil, errors.Errorf("rowsFrom: row %d column mismatch", index)
		}
		prepared[index] = item
	}

	return prepared, nil
}

func (this *PgQueryBuilder) rowFromMap(
	schema *dmodel.ModelSchema, values dmodel.DynamicFields, include func(string) bool,
) (rowData, error) {
	includeFn := include
	if includeFn == nil {
		includeFn = func(string) bool { return true }
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if !includeFn(key) {
			continue
		}
		field, ok := schema.Column(key)
		if !ok {
			return rowData{}, errors.Wrap(&errClientUnknownField{Field: key}, "rowFromMap")
		}
		if field.IsVirtualModelField() {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := rowData{
		columns: keys,
		values:  make([]any, len(keys)),
	}
	for i, key := range keys {
		field, ok := schema.Column(key)
		if !ok {
			return rowData{}, errors.Wrap(&errClientUnknownField{Field: key}, "rowFromMap")
		}
		converted, err := this.convertValue(field, values[key])
		if err != nil {
			return rowData{}, errors.Wrapf(err, "rowFromMap: invalid value for column '%s'", key)
		}
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) rowForKeys(schema *dmodel.ModelSchema, values dmodel.DynamicFields, keys []string) (rowData, error) {
	result := rowData{
		columns: make([]string, len(keys)),
		values:  make([]any, len(keys)),
	}

	for i, key := range keys {
		field, ok := schema.Column(key)
		if !ok {
			return rowData{}, errors.Wrap(&errClientUnknownField{Field: key}, "rowForKeys")
		}
		if field.IsVirtualModelField() {
			return rowData{}, errors.Errorf("rowForKeys: key field '%s' is not defined in this schema", key)
		}
		raw, ok := values[key]
		if !ok {
			return rowData{}, errors.Errorf("rowForKeys: missing key '%s'", key)
		}
		converted, err := this.convertValue(field, raw)
		if err != nil {
			return rowData{}, errors.Wrapf(err, "rowForKeys: invalid value for key '%s'", key)
		}
		result.columns[i] = key
		result.values[i] = converted
	}

	return result, nil
}

func (this *PgQueryBuilder) graphExpression(
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	condition *dmodel.Condition,
	and []dmodel.SearchNode,
	or []dmodel.SearchNode,
) (condStr string, ok bool, err error) {
	switch {
	case condition != nil:
		condStr, err = this.conditionExpression(schema, sb, *condition)
		return condStr, err == nil, err
	case len(and) > 0:
		return this.combineNodes(schema, sb, and, sb.And)
	case len(or) > 0:
		return this.combineNodes(schema, sb, or, sb.Or)
	default:
		return "", false, nil
	}
}

func (this *PgQueryBuilder) combineNodes(
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	nodes []dmodel.SearchNode,
	join func(...string) string,
) (string, bool, error) {
	conditions := make([]string, 0, len(nodes))
	for _, node := range nodes {
		condStr, condOk, condErr := this.graphExpression(
			schema, sb, node.GetCondition(), node.GetAnd(), node.GetOr())
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
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	cond dmodel.Condition,
) (string, error) {
	originalField := cond.Field()
	fieldName := originalField
	operator := cond.Operator()
	value := cond.Value()
	valueArr := cond.Values()

	predefinedPredicateFn := this.GetPredefinedPredicate(fieldName, schema.Name())
	if predefinedPredicateFn != nil {
		result, cErr := predefinedPredicateFn(operator, value)
		if cErr.Count() > 0 {
			return "", wrapClientSqlErrors(cErr)
		}
		if result.NewFieldName != "" {
			fieldName = result.NewFieldName
		}
		if result.NewOperator != "" {
			operator = result.NewOperator
		}
		if result.NewValue != nil {
			value = result.NewValue
		}
		if result.NewValues != nil {
			valueArr = result.NewValues
		}
	}

	field, quotedField, err := this.prepareColName(schema, fieldName)
	if err != nil {
		return "", err
	}

	switch operator {
	case dmodel.Equals, dmodel.NotEquals, dmodel.GreaterThan,
		dmodel.GreaterEqual, dmodel.LessThan, dmodel.LessEqual:
		return this.comparisonPredicate(sb, quotedField, field, operator, value)
	case dmodel.In, dmodel.NotIn:
		return this.collectionPredicate(sb, quotedField, field, operator, valueArr)
	case dmodel.Contains, dmodel.NotContains, dmodel.StartsWith,
		dmodel.NotStartsWith, dmodel.EndsWith, dmodel.NotEndsWith:
		return this.stringPredicate(sb, quotedField, field, operator, value)
	case dmodel.IsSet, dmodel.IsNotSet:
		return nullPredicate(sb, quotedField, operator), nil
	default:
		return "", wrapClientSqlErrors(ClientErrorsUnsupportedFilterOperator(originalField))
	}
}

func (this *PgQueryBuilder) prepareColName(
	schema *dmodel.ModelSchema, fieldName string,
) (*dmodel.ModelField, string, error) {
	if strings.Contains(fieldName, ".") {
		return nil, "", wrapClientSqlErrors(clientErrorsNestedFieldNotSupported(fieldName))
	}
	field, ok := schema.Column(fieldName)
	if !ok || field.IsVirtualModelField() {
		return nil, "", errors.Wrap(&errClientUnknownField{Field: fieldName}, "prepareColName")
	}
	return field, pgQuote(field.Name()), nil
}

func (this *PgQueryBuilder) comparisonPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *dmodel.ModelField, op dmodel.Operator, value any,
) (string, error) {
	converted, err := this.convertValue(field, value)
	if err != nil {
		return "", err
	}
	switch op {
	case dmodel.Equals:
		return sb.Equal(quotedField, converted), nil
	case dmodel.NotEquals:
		return sb.NotEqual(quotedField, converted), nil
	case dmodel.GreaterThan:
		return sb.GreaterThan(quotedField, converted), nil
	case dmodel.GreaterEqual:
		return sb.GreaterEqualThan(quotedField, converted), nil
	case dmodel.LessThan:
		return sb.LessThan(quotedField, converted), nil
	case dmodel.LessEqual:
		return sb.LessEqualThan(quotedField, converted), nil
	default:
		panic("comparisonPredicate: unsupported operator (internal)")
	}
}

func (this *PgQueryBuilder) collectionPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *dmodel.ModelField, op dmodel.Operator, values []any,
) (string, error) {
	converted, err := this.convertValues(field, values)
	if err != nil {
		return "", err
	}
	if op == dmodel.In {
		return sb.In(quotedField, converted...), nil
	}
	if op == dmodel.NotIn {
		return sb.NotIn(quotedField, converted...), nil
	}
	return "", errors.Errorf("collectionPredicate: unsupported collection operator '%s'", op)
}

func (this *PgQueryBuilder) stringPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *dmodel.ModelField, op dmodel.Operator, value any,
) (string, error) {
	if columnCategoryFor(field.ColumnType()) != columnString {
		return "", errors.Errorf(
			"stringPredicate: operator '%s' requires string field '%s'", op, field.Name())
	}
	converted, err := this.convertValue(field, value)
	if err != nil {
		return "", err
	}
	pattern := stringPattern(fmt.Sprint(converted), op)
	if pattern == "" {
		return "", errors.Errorf("stringPredicate: unsupported string operator '%s'", op)
	}
	switch op {
	case dmodel.NotContains, dmodel.NotStartsWith, dmodel.NotEndsWith:
		return sb.NotILike(quotedField, pattern), nil
	default:
		return sb.ILike(quotedField, pattern), nil
	}
}

func nullPredicate(sb *sqlbuilder.SelectBuilder, quotedField string, op dmodel.Operator) string {
	if op == dmodel.IsSet {
		return sb.IsNotNull(quotedField)
	}
	return sb.IsNull(quotedField)
}

func (this *PgQueryBuilder) convertValues(field *dmodel.ModelField, values []any) ([]any, error) {
	if len(values) == 0 {
		return nil, errors.Errorf(
			"convertValues: operator requires at least one value for field '%s'", field.Name())
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
	schema *dmodel.ModelSchema, order dmodel.SearchOrder,
) ([]string, error) {
	exprs := make([]string, 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		fieldName := item[0]
		if strings.Contains(fieldName, ".") {
			return nil, wrapClientSqlErrors(clientErrorsNestedFieldNotSupported(fieldName))
		}
		field, ok := schema.Column(fieldName)
		if !ok {
			return nil, errors.Wrap(&errClientUnknownField{Field: fieldName}, "orderExprs")
		}
		if field.IsVirtualModelField() {
			return nil, errors.Errorf(
				"orderExprs: order field '%s' is not stored in this schema", fieldName)
		}
		dir := "ASC"
		if item.Direction() == dmodel.Desc {
			dir = "DESC"
		}
		exprs = append(exprs, fmt.Sprintf("%s %s", pgQuote(fieldName), dir))
	}
	return exprs, nil
}

func (this *PgQueryBuilder) convertValue(field *dmodel.ModelField, value any) (any, error) {
	if field.IsVirtualModelField() {
		return nil, errors.Errorf("convertValue: field '%s' is not available", field.Name())
	}
	if value == nil {
		if field.IsNullable() {
			return nil, nil
		}
		return nil, errors.Errorf("convertValue: field '%s' does not allow NULL", field.Name())
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if field.IsNullable() {
			return nil, nil
		}
		return nil, errors.Errorf("convertValue: field '%s' does not allow NULL", field.Name())
	}
	if !valueAllowed(columnCategoryFor(field.ColumnType()), v) {
		return nil, errors.Errorf(
			"convertValue: field '%s': incompatible value type %T", field.Name(), v.Interface())
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
		return "", errors.Errorf("resolveGenericToPgType: unsupported generic type '%s'", genericType)
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

func stringPattern(value string, op dmodel.Operator) string {
	switch op {
	case dmodel.Contains, dmodel.NotContains:
		return "%" + value + "%"
	case dmodel.StartsWith, dmodel.NotStartsWith:
		return value + "%"
	case dmodel.EndsWith, dmodel.NotEndsWith:
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
		strings.Contains(typ, "date"), strings.Contains(typ, "time"), typ == "nikkiDate", typ == "nikkiTime", typ == "nikkiDateTime":
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
var modelDateType = reflect.TypeOf(cmodel.ModelDate(time.Time{}))
var modelDateTimeType = reflect.TypeOf(cmodel.ModelDateTime(time.Time{}))
var modelTimeType = reflect.TypeOf(cmodel.ModelTime(time.Time{}))

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
		return v.Type() == timeType ||
			v.Type() == modelDateType ||
			v.Type() == modelDateTimeType ||
			v.Type() == modelTimeType
	case columnJSON:
		return v.Kind() == reflect.Map
	default:
		return true
	}
}

type errClientUnknownField struct {
	Field string
}

func (e *errClientUnknownField) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("unknown field '%s'", e.Field)
}

func clientErrorsUnknownField(field string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_unknown_schema_field"),
			"field is not defined on this schema"),
	}
}

func extractClientErr[T any](err error) (*crud.OpResult[T], error) {
	if err == nil {
		return nil, nil
	}
	var unk *errClientUnknownField
	if stdErr.As(err, &unk) {
		return &crud.OpResult[T]{ClientErrors: clientErrorsUnknownField(unk.Field)}, nil
	}
	var sqlErr *errClientSqlErrors
	if stdErr.As(err, &sqlErr) {
		return &crud.OpResult[T]{ClientErrors: sqlErr.errors}, nil
	}
	return nil, err
}
