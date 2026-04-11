package orm

import (
	"encoding/json"
	stdErr "errors"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/convert"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
)

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
) ([]string, *ft.ClientErrors, error) {
	createSql, err := this.buildCreateTableSql(schema, registry)
	if err != nil {
		return nil, nil, err
	}
	indexSqls, err := this.partialUniqueIndexSqls(schema)
	if err != nil {
		return nil, nil, err
	}
	searchIndexSqls, err := this.searchIndexSqls(schema)
	if err != nil {
		return nil, nil, err
	}
	out := append([]string{createSql}, indexSqls...)
	out = append(out, searchIndexSqls...)
	return out, nil, nil
}

func (this *PgQueryBuilder) buildCreateTableSql(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) (string, error) {
	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder().CreateTable(pgQuote(schema.TableName()))
	if err := this.defineColumns(builder, schema); err != nil {
		return "", err
	}
	this.defineKeys(builder, schema)
	if err := this.defineForeignKeys(builder, schema, registry); err != nil {
		return "", err
	}
	sql, _ := builder.Build()
	return strings.TrimSuffix(sql, ";"), nil
}

func (this *PgQueryBuilder) partialUniqueIndexSqls(schema *dmodel.ModelSchema) ([]string, error) {
	groups := schema.PartialUniqueGroups()
	out := make([]string, 0, len(groups)*2)
	tenantKey := schema.TenantKey()
	for _, group := range groups {
		group.NotNullFields = prependTenantKey(tenantKey, group.NotNullFields)
		lines, err := formatPartialUniqueGroupIndexPair(schema, group)
		if err != nil {
			return nil, err
		}
		out = append(out, lines...)
	}
	return out, nil
}

func (this *PgQueryBuilder) searchIndexSqls(schema *dmodel.ModelSchema) ([]string, error) {
	groups := schema.SearchIndexGroups()
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		line, err := formatSearchIndexGroup(schema, group)
		if err != nil {
			return nil, err
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}

func schemaHasSingleColumnUniqueOn(schema *dmodel.ModelSchema, col string) bool {
	for _, u := range schema.AllUniques() {
		if len(u) == 1 && u[0] == col {
			return true
		}
	}
	return false
}

func formatPartialUniqueGroupIndexPair(schema *dmodel.ModelSchema, group dmodel.PartialUniqueGroupParam) ([]string, error) {
	if len(group.NotNullFields) == 0 {
		return nil, errors.Errorf(
			"formatPartialUniqueGroupIndexPair: table '%s': at least one not-null field is required", schema.TableName())
	}
	nullable := strings.TrimSpace(group.NullableField)
	if nullable == "" {
		return nil, errors.Errorf(
			"formatPartialUniqueGroupIndexPair: table '%s': nullable field is required", schema.TableName())
	}
	for _, col := range group.NotNullFields {
		if schemaHasSingleColumnUniqueOn(schema, col) {
			return nil, errors.Errorf(
				"partialUniqueIndexSqls: table '%s': column '%s' already has a single-column UNIQUE constraint",
				schema.TableName(), col)
		}
	}
	indexName := resolvePartialUniqueGroupIndexName(schema.TableName(), group)
	colsWithNullable := make([]string, 0, len(group.NotNullFields)+1)
	for _, col := range group.NotNullFields {
		colsWithNullable = append(colsWithNullable, pgQuote(col))
	}
	colsWithNullable = append(colsWithNullable, pgQuote(nullable))
	quotedNotNull := pgQuoteArr(group.NotNullFields)
	tableRef := pgQuoteTable(strings.Split(schema.TableName(), ".")...)
	lineNN := fmt.Sprintf(
		"CREATE UNIQUE INDEX %s ON %s (%s) WHERE %s IS NOT NULL",
		pgQuote(indexName+"_ukey_notnull"),
		tableRef,
		strings.Join(colsWithNullable, ", "),
		pgQuote(nullable),
	)
	lineNull := fmt.Sprintf(
		"CREATE UNIQUE INDEX %s ON %s (%s) WHERE %s IS NULL",
		pgQuote(indexName+"_ukey_null"),
		tableRef,
		strings.Join(quotedNotNull, ", "),
		pgQuote(nullable),
	)
	return []string{lineNN, lineNull}, nil
}

func resolvePartialUniqueGroupIndexName(tableName string, group dmodel.PartialUniqueGroupParam) string {
	raw := strings.TrimSpace(group.IndexName)
	if raw == "" {
		raw = fmt.Sprintf("%s_%s_%s", tableName, strings.Join(group.NotNullFields, "_"), group.NullableField)
	}
	return toSnakeLower(raw)
}

func toSnakeLower(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	var b strings.Builder
	runes := []rune(strings.TrimSpace(input))
	for i, current := range runes {
		if current == ' ' || current == '-' || current == '.' {
			if b.Len() > 0 && b.String()[b.Len()-1] != '_' {
				b.WriteByte('_')
			}
			continue
		}
		if i > 0 && isUpperAscii(current) && (isLowerOrDigitAscii(runes[i-1]) ||
			(i+1 < len(runes) && isLowerAscii(runes[i+1]))) {
			if b.Len() > 0 && b.String()[b.Len()-1] != '_' {
				b.WriteByte('_')
			}
		}
		b.WriteRune(current)
	}
	return strings.ToLower(convert.ToUnicodeSnakeCase(b.String()))
}

func formatSearchIndexGroup(schema *dmodel.ModelSchema, group dmodel.SearchIndexGroupParam) (string, error) {
	if len(group.Fields) == 0 {
		return "", nil
	}
	indexName := resolveSearchIndexGroupIndexName(group)
	if indexName == "" {
		return "", errors.Errorf(
			"formatSearchIndexGroup: table '%s': index name is required", schema.TableName())
	}
	tableRef := pgQuoteTable(strings.Split(schema.TableName(), ".")...)
	return fmt.Sprintf(
		"CREATE INDEX %s ON %s (%s)",
		pgQuote(indexName),
		tableRef,
		strings.Join(pgQuoteArr(group.Fields), ", "),
	), nil
}

func resolveSearchIndexGroupIndexName(group dmodel.SearchIndexGroupParam) string {
	raw := strings.TrimSpace(group.IndexName)
	if raw == "" {
		raw = strings.Join(group.Fields, "_") + "_idx"
	}
	return toSnakeLower(raw)
}

func isUpperAscii(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLowerAscii(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isLowerOrDigitAscii(r rune) bool {
	return isLowerAscii(r) || (r >= '0' && r <= '9')
}

func (this *PgQueryBuilder) defineColumns(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema,
) error {
	for _, col := range schema.Columns() {
		pgType, err := resolveModelFieldToPgType(col)
		if err != nil {
			return errors.Wrapf(err, "defineColumns: column '%s'", col.Name())
		}
		builder.Define(pgQuote(col.Name()), pgType, col.ColumnNullable())
	}
	return nil
}

func (this *PgQueryBuilder) defineKeys(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema,
) {
	if keys := schema.PrimaryKeys(); len(keys) > 0 {
		builder.Define("PRIMARY KEY", fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(keys), ", ")))
	}
	tenantKey := schema.TenantKey()
	for _, unique := range schema.AllUniques() {
		if len(unique) == 0 {
			continue
		}
		effectiveUnique := prependTenantKey(tenantKey, unique)
		name := pgQuote(fmt.Sprintf("%s_%s_ukey", schema.TableName(), strings.Join(effectiveUnique, "_")))
		cols := fmt.Sprintf("(%s)", strings.Join(pgQuoteArr(effectiveUnique), ", "))
		builder.Define(fmt.Sprintf("CONSTRAINT %s UNIQUE", name), cols)
	}
}

func (this *PgQueryBuilder) defineForeignKeys(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) error {
	if err := this.defineForeignKeysDeclaredOnSchema(builder, schema, registry); err != nil {
		return err
	}
	return this.defineForeignKeysFromParentOneToMany(builder, schema, registry)
}

func (this *PgQueryBuilder) defineForeignKeysDeclaredOnSchema(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) error {
	for _, rel := range schema.ToRelations() {
		if !isFkOwnerRelationType(rel.RelationType) {
			continue
		}
		refSchema := registry.Get(rel.DestSchemaName)
		if refSchema == nil {
			return errors.Errorf(
				"defineForeignKeys: referenced schema not found for relation '%s'", rel.Edge,
			)
		}
		if err := this.appendCompositeForeignKey(builder, schema, refSchema, rel); err != nil {
			return err
		}
	}
	return nil
}

func (this *PgQueryBuilder) defineForeignKeysFromParentOneToMany(
	builder *sqlbuilder.CreateTableBuilder, schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry,
) error {
	return registry.ForEach(func(_ string, parentSch *dmodel.ModelSchema) error {
		for _, rel := range parentSch.ToRelations() {
			if rel.RelationType != dmodel.RelationTypeOneToMany || rel.DestSchemaName != schema.Name() {
				continue
			}
			if schemaAlreadyDeclaresFkToParent(schema, parentSch, rel) {
				continue
			}
			if err := this.appendCompositeForeignKey(builder, schema, parentSch, rel); err != nil {
				return err
			}
		}
		return nil
	})
}

func schemaAlreadyDeclaresFkToParent(
	child, parent *dmodel.ModelSchema, parentOneToMany dmodel.ModelRelation,
) bool {
	for _, r := range child.ToRelations() {
		if !isFkOwnerRelationType(r.RelationType) || r.DestSchemaName != parent.Name() {
			continue
		}
		if dmodel.RelationsShareForeignKeyColumns(r, parentOneToMany) {
			return true
		}
	}
	return false
}

func (this *PgQueryBuilder) appendCompositeForeignKey(
	builder *sqlbuilder.CreateTableBuilder,
	fkOwner *dmodel.ModelSchema,
	refSchema *dmodel.ModelSchema,
	rel dmodel.ModelRelation,
) error {
	pairs := rel.EffectiveForeignKeys()
	if len(pairs) == 0 {
		return errors.Errorf("appendCompositeForeignKey: relation '%s' has no FK columns", rel.Edge)
	}
	fkCols := make([]string, len(pairs))
	refCols := make([]string, len(pairs))
	for i, p := range pairs {
		fkCols[i] = pgQuote(p.FkColumn)
		refCols[i] = pgQuote(p.ReferencedColumn)
	}
	suffix := pairs[0].FkColumn
	if len(pairs) > 1 {
		suffix = fmt.Sprintf("comp_%s", rel.Edge)
	}
	fkName := pgQuote(fmt.Sprintf("%s_%s_fkey", fkOwner.TableName(), suffix))
	fkBody := fmt.Sprintf("(%s) REFERENCES %s (%s) ON UPDATE %s ON DELETE %s",
		strings.Join(fkCols, ", "), pgQuote(refSchema.TableName()), strings.Join(refCols, ", "),
		rel.OnUpdate.Sql(), rel.OnDelete.Sql())
	builder.Define(fmt.Sprintf("CONSTRAINT %s FOREIGN KEY", fkName), fkBody)
	return nil
}

func (this *PgQueryBuilder) SqlSelectGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildSqlSelectGraph(schema, registry, graph, opts)
	return stringSqlGraphOutcome(sql, cErrs, err)
}

func (this *PgQueryBuilder) buildSqlSelectGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (string, ft.ClientErrors, error) {
	planner, err := this.planGraphJoins(schema, registry, graph, opts)
	if err != nil {
		return "", nil, err
	}
	ctx := &graphSelectCtx{planner: planner, language: opts.Language}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	if anySelectColumnDistinct(opts.Columns) {
		sb.Distinct()
	}
	if err := this.applySelectColumns(sb, planner, opts.Columns); err != nil {
		return "", nil, err
	}
	this.applyFromWithJoins(sb, schema, planner)
	this.appendPlannerM2MTenantWheres(sb, planner)
	if graph != nil {
		predicate, graphCErrs, err := this.graphExpression(
			ctx, schema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
		if err != nil {
			return "", nil, err
		}
		if len(graphCErrs) > 0 {
			return "", graphCErrs, nil
		}
		if len(predicate) > 0 {
			sb.Where(predicate)
		}
		orderExprs, err := this.orderExprs(ctx, schema, graph.GetOrder())
		if err != nil {
			return "", nil, err
		}
		if len(orderExprs) > 0 {
			sb.OrderBy(orderExprs...)
		}
	}
	this.applyPagination(sb, opts.Page, opts.Size)
	sql, args := sb.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", nil, errors.Wrap(ierr, "buildSqlSelectGraph: interpolate")
	}
	return out, nil, nil
}

func (this *PgQueryBuilder) SqlExistsGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildSqlExistsGraph(schema, registry, graph)
	return stringSqlGraphOutcome(sql, cErrs, err)
}

func (this *PgQueryBuilder) buildSqlExistsGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph,
) (string, ft.ClientErrors, error) {
	planner, err := this.planGraphJoins(schema, registry, graph, SqlSelectGraphOpts{})
	if err != nil {
		return "", nil, err
	}
	ctx := &graphSelectCtx{planner: planner}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("1")
	this.applyFromWithJoins(sb, schema, planner)
	this.appendPlannerM2MTenantWheres(sb, planner)
	if graph != nil {
		predicate, graphCErrs, err := this.graphExpression(
			ctx, schema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
		if err != nil {
			return "", nil, err
		}
		if len(graphCErrs) > 0 {
			return "", graphCErrs, nil
		}
		if len(predicate) > 0 {
			sb.Where(predicate)
		}
	}
	innerSql, args := sb.Build()
	innerOut, ierr := interpolate(innerSql, args)
	if ierr != nil {
		return "", nil, errors.Wrap(ierr, "buildSqlExistsGraph: interpolate inner")
	}
	return fmt.Sprintf("SELECT EXISTS (%s)", innerOut), nil, nil
}

func (this *PgQueryBuilder) SqlCountGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildSqlCountGraph(schema, registry, graph, opts)
	return stringSqlGraphOutcome(sql, cErrs, err)
}

func (this *PgQueryBuilder) buildSqlCountGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (string, ft.ClientErrors, error) {
	planner, err := this.planGraphJoins(schema, registry, graph, opts)
	if err != nil {
		return "", nil, err
	}
	ctx := &graphSelectCtx{planner: planner, language: opts.Language}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("COUNT(*)")
	this.applyFromWithJoins(sb, schema, planner)
	this.appendPlannerM2MTenantWheres(sb, planner)
	if graph != nil {
		predicate, graphCErrs, err := this.graphExpression(
			ctx, schema, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
		if err != nil {
			return "", nil, err
		}
		if len(graphCErrs) > 0 {
			return "", graphCErrs, nil
		}
		if len(predicate) > 0 {
			sb.Where(predicate)
		}
	}
	sql, args := sb.Build()
	out, ierr := interpolate(sql, args)
	if ierr != nil {
		return "", nil, errors.Wrap(ierr, "buildSqlCountGraph: interpolate")
	}
	return out, nil, nil
}

func (this *PgQueryBuilder) appendPlannerM2MTenantWheres(sb *sqlbuilder.SelectBuilder, planner *joinPlanner) {
	if planner == nil {
		return
	}
	for _, w := range planner.m2mTenantWheres {
		sb.Where(w)
	}
}

func (this *PgQueryBuilder) applyFromWithJoins(
	sb *sqlbuilder.SelectBuilder, schema *dmodel.ModelSchema, planner *joinPlanner,
) {
	if planner == nil || !planner.usesJoins() {
		sb.From(this.tableExpression(schema))
		return
	}
	planner.ensureRootAliased()
	sb.From(fmt.Sprintf("%s AS %s", this.tableExpression(schema), planner.rootAlias))
	for _, j := range planner.joins {
		sb.Join(j.tableWithAlias, j.onExpr)
	}
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

func (this *PgQueryBuilder) SqlInsert(schema *dmodel.ModelSchema, data dmodel.DynamicFields,
	ignoreConflict bool,
) (*string, *ft.ClientErrors, error) {
	return this.SqlInsertBulk(schema, []dmodel.DynamicFields{data}, ignoreConflict)
}

func (this *PgQueryBuilder) SqlInsertBulk(schema *dmodel.ModelSchema, rows []dmodel.DynamicFields,
	ignoreConflict bool,
) (*string, *ft.ClientErrors, error) {
	prepared, cErrs, err := this.rowsFrom(schema, rows, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(cErrs) > 0 {
		return nil, clientErrsPtr(cErrs), nil
	}
	sql, ierr := this.buildInsertSql(schema, prepared, ignoreConflict)
	return stringSqlOutcome(sql, ierr)
}

func (this *PgQueryBuilder) appendInsertOnConflictPkDoNothing(ib *sqlbuilder.InsertBuilder, schema *dmodel.ModelSchema) error {
	pks := schema.PrimaryKeys()
	if len(pks) == 0 {
		return errors.New("appendInsertOnConflictPkDoNothing: schema has no primary keys")
	}
	quoted := pgQuoteArr(pks)
	ib.SQL(fmt.Sprintf(" ON CONFLICT (%s) DO NOTHING", strings.Join(quoted, ", ")))
	return nil
}

func (this *PgQueryBuilder) buildInsertSql(schema *dmodel.ModelSchema, rows []rowData, ignoreConflict bool) (
	string, error,
) {
	if len(rows) == 0 {
		return "", errors.New("buildInsertSql: no rows provided")
	}
	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
	ib.InsertInto(this.tableExpression(schema))
	ib.Cols(pgQuoteArr(rows[0].columns)...)
	for _, row := range rows {
		ib.Values(row.values...)
	}
	if ignoreConflict {
		ib.SQL(" ON CONFLICT DO NOTHING")
		// if err := this.appendInsertOnConflictPkDoNothing(ib, schema); err != nil {
		// 	return "", err
		// }
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
) (*string, *ft.ClientErrors, error) {
	if len(filters) == 0 {
		return nil, nil, errors.New("SqlUpdateEqual: no filters provided")
	}

	target, cErrs, err := this.rowFromMap(schema, data, func(name string) bool {
		return !schema.IsPrimaryKey(name) && !schema.IsTenantKey(name)
	})
	if err != nil {
		return nil, nil, err
	}
	if len(cErrs) > 0 {
		return nil, clientErrsPtr(cErrs), nil
	}
	if len(target.columns) == 0 {
		return nil, nil, errors.New("SqlUpdateEqual: no updatable columns provided")
	}

	lookup, lookupCErrs, err := this.rowFromMap(schema, filters, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(lookupCErrs) > 0 {
		return nil, clientErrsPtr(lookupCErrs), nil
	}
	sql, ierr := this.buildUpdateSql(schema, target, lookup)
	return stringSqlOutcome(sql, ierr)
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
	*string, *ft.ClientErrors, error,
) {
	if len(filters) == 0 {
		return nil, nil, errors.New("SqlDeleteEqual: no filters provided")
	}

	row, cErrs, err := this.rowFromMap(schema, filters, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(cErrs) > 0 {
		return nil, clientErrsPtr(cErrs), nil
	}
	if len(row.columns) == 0 {
		return nil, nil, errors.New("SqlDeleteEqual: no filters provided")
	}

	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	db.DeleteFrom(this.tableExpression(schema))
	for i, col := range row.columns {
		db.Where(db.Equal(pgQuote(col), row.values[i]))
	}
	sql, args := db.Build()
	out, ierr := interpolate(sql, args)
	var interpErr error
	if ierr != nil {
		interpErr = errors.Wrap(ierr, "SqlDeleteEqual: interpolate")
	}
	return stringSqlOutcome(out, interpErr)
}

func (this *PgQueryBuilder) SqlDeleteOrAndEquals(
	schema *dmodel.ModelSchema, filters []dmodel.DynamicFields,
) (*string, *ft.ClientErrors, error) {
	if len(filters) == 0 {
		return nil, nil, errors.New("SqlDeleteOrAndEquals: no filters provided")
	}
	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	db.DeleteFrom(this.tableExpression(schema))
	orClauses := make([]string, 0, len(filters))
	for _, f := range filters {
		row, rowCErrs, err := this.rowFromMap(schema, f, nil)
		if err != nil {
			return nil, nil, err
		}
		if len(rowCErrs) > 0 {
			return nil, clientErrsPtr(rowCErrs), nil
		}
		if len(row.columns) == 0 {
			return nil, nil, errors.New("SqlDeleteOrAndEquals: empty filter conjunct")
		}
		ands := make([]string, 0, len(row.columns))
		for i, col := range row.columns {
			quoted := pgQuote(col)
			if row.values[i] == nil {
				ands = append(ands, db.IsNull(quoted))
			} else {
				ands = append(ands, db.Equal(quoted, row.values[i]))
			}
		}
		orClauses = append(orClauses, db.And(ands...))
	}
	db.Where(db.Or(orClauses...))
	sql, args := db.Build()
	out, ierr := interpolate(sql, args)
	return stringSqlOutcome(out, ierr)
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
) (*SqlCheckUniqueCollisionsData, *ft.ClientErrors, error) {
	if len(uniqueKeysToCheck) == 0 {
		return nil, nil, nil
	}

	tableRef := this.tableExpression(schema)
	tenantKey := schema.TenantKey()
	var args []any
	argIdx := 1
	var parts []string

	for _, uniqueFields := range uniqueKeysToCheck {
		part, partArgs, partCErrs, err := this.buildUniqueCheckPart(schema, tableRef, tenantKey, uniqueFields, data, argIdx)
		if err != nil {
			return nil, nil, err
		}
		if len(partCErrs) > 0 {
			return nil, clientErrsPtr(partCErrs), nil
		}
		parts = append(parts, part)
		args = append(args, partArgs...)
		argIdx += len(partArgs)
	}

	joined := strings.Join(parts, " UNION ALL ")
	out := SqlCheckUniqueCollisionsData{Sql: joined, Args: args}
	return &out, nil, nil
}

func (this *PgQueryBuilder) SqlExistsMany(
	schema *dmodel.ModelSchema, keys []dmodel.DynamicFields,
) (*SqlExistsManyData, *ft.ClientErrors, error) {
	if len(keys) == 0 {
		return nil, nil, nil
	}
	tableRef := this.tableExpression(schema)
	prepared, cErrs, err := this.rowsFrom(schema, keys, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(cErrs) > 0 {
		return nil, clientErrsPtr(cErrs), nil
	}
	var args []any
	parts := make([]string, 0, len(prepared))
	argIdx := 1
	for _, row := range prepared {
		part := buildExistsCaseSql(tableRef, row.columns, argIdx)
		parts = append(parts, part)
		args = append(args, row.values...)
		argIdx += len(row.values)
	}
	joined := strings.Join(parts, " UNION ALL ")
	out := SqlExistsManyData{Sql: joined, Args: args}
	return &out, nil, nil
}

func (this *PgQueryBuilder) buildUniqueCheckPart(
	schema *dmodel.ModelSchema, tableRef string, tenantKey string,
	uniqueFields []string, data dmodel.DynamicFields, argIdx int,
) (string, []any, ft.ClientErrors, error) {
	if len(uniqueFields) == 0 {
		return "SELECT 0", nil, nil, nil
	}
	columns := prependTenantKey(tenantKey, uniqueFields)
	values, hasAll, cErrs, err := this.resolveColumnValues(schema, columns, data)

	if err != nil {
		return "", nil, nil, err
	}
	if len(cErrs) > 0 {
		return "", nil, cErrs, nil
	}
	if !hasAll {
		return "SELECT 0", nil, nil, nil
	}

	ignoreCurrentRowCond := this.buildIgnoreCurrentRowCond(schema, data)
	if ignoreCurrentRowCond != "" {
		return buildExistsCaseSqlWithOptCond(tableRef, columns, argIdx, ignoreCurrentRowCond), values, nil, nil
	}

	return buildExistsCaseSql(tableRef, columns, argIdx), values, nil, nil
}

func (this *PgQueryBuilder) buildIgnoreCurrentRowCond(schema *dmodel.ModelSchema, data dmodel.DynamicFields) string {
	primaryKeys := schema.PrimaryKeys()
	if len(primaryKeys) == 0 {
		return ""
	}

	conds := make([]string, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		v, ok := data[key]
		if !ok || v == nil {
			return ""
		}
		conds = append(conds, fmt.Sprintf("%s <> %v", pgQuote(key), v))
	}

	return strings.Join(conds, " OR ")
}

func (this *PgQueryBuilder) resolveColumnValues(
	schema *dmodel.ModelSchema, columns []string, data dmodel.DynamicFields,
) ([]any, bool, ft.ClientErrors, error) {
	values := make([]any, 0, len(columns))
	for _, col := range columns {
		v, ok := data[col]
		if !ok || v == nil {
			return nil, false, nil, nil
		}
		field, ok := schema.Column(col)
		if !ok || field.IsVirtualModelField() {
			return nil, false, nil, errors.Wrap(&errClientUnknownField{Field: col}, "resolveColumnValues")
		}
		converted, cErrs, err := this.convertValue(field, v)
		if err != nil {
			return nil, false, nil, errors.Wrapf(err, "resolveColumnValues: column '%s'", col)
		}
		if len(cErrs) > 0 {
			return nil, false, cErrs, nil
		}
		values = append(values, converted)
	}
	return values, true, nil, nil
}

type rowData struct {
	columns []string
	values  []any
}

func (this *PgQueryBuilder) rowsFrom(
	schema *dmodel.ModelSchema, rows []dmodel.DynamicFields, filter func(string) bool,
) ([]rowData, ft.ClientErrors, error) {
	if len(rows) == 0 {
		return nil, nil, errors.New("rowsFrom: no rows provided")
	}

	prepared := make([]rowData, len(rows))
	var reference []string

	for index, row := range rows {
		item, cErrs, err := this.rowFromMap(schema, row, filter)
		if err != nil {
			return nil, nil, err
		}
		if len(cErrs) > 0 {
			return nil, cErrs, nil
		}
		if len(item.columns) == 0 {
			return nil, nil, errors.New("rowsFrom: no columns provided")
		}
		if index == 0 {
			reference = item.columns
		} else if !slices.Equal(reference, item.columns) {
			return nil, nil, errors.Errorf("rowsFrom: row %d column mismatch", index)
		}
		prepared[index] = item
	}

	return prepared, nil, nil
}

func (this *PgQueryBuilder) rowFromMap(
	schema *dmodel.ModelSchema, values dmodel.DynamicFields, include func(string) bool,
) (rowData, ft.ClientErrors, error) {
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
			return rowData{}, nil, errors.Wrap(&errClientUnknownField{Field: key}, "rowFromMap")
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
			return rowData{}, nil, errors.Wrap(&errClientUnknownField{Field: key}, "rowFromMap")
		}
		converted, cErrs, err := this.convertValue(field, values[key])
		if err != nil {
			return rowData{}, nil, errors.Wrapf(err, "rowFromMap: invalid value for column '%s'", key)
		}
		if len(cErrs) > 0 {
			return rowData{}, cErrs, nil
		}
		result.values[i] = converted
	}

	return result, nil, nil
}

func (this *PgQueryBuilder) graphExpression(
	ctx *graphSelectCtx,
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	condition dmodel.Condition,
	and []dmodel.SearchNode,
	or []dmodel.SearchNode,
) (condStr string, cErrs ft.ClientErrors, err error) {
	switch {
	case condition.Field() != "":
		condStr, cErrs, err = this.conditionExpression(ctx, schema, sb, condition)
		return condStr, cErrs, err
	case len(and) > 0:
		return this.combineNodes(ctx, schema, sb, and, sb.And)
	case len(or) > 0:
		return this.combineNodes(ctx, schema, sb, or, sb.Or)
	default:
		return "", nil, nil
	}
}

func (this *PgQueryBuilder) combineNodes(
	ctx *graphSelectCtx,
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	nodes []dmodel.SearchNode,
	join func(...string) string,
) (string, ft.ClientErrors, error) {
	conditions := make([]string, 0, len(nodes))
	for _, node := range nodes {
		condStr, nodeCErrs, condErr := this.graphExpression(
			ctx, schema, sb, node.GetCondition(), node.GetAnd(), node.GetOr())
		if condErr != nil {
			return "", nil, condErr
		}
		if len(nodeCErrs) > 0 {
			return "", nodeCErrs, nil
		}
		if len(condStr) > 0 {
			conditions = append(conditions, condStr)
		}
	}
	if len(conditions) == 0 {
		return "", nil, nil
	}
	return join(conditions...), nil, nil
}

func (this *PgQueryBuilder) conditionExpression(
	ctx *graphSelectCtx,
	schema *dmodel.ModelSchema,
	sb *sqlbuilder.SelectBuilder,
	cond dmodel.Condition,
) (string, ft.ClientErrors, error) {
	originalField := cond.Field()
	fieldName := originalField
	operator := cond.Operator()
	value := derefConditionOperand(cond.Value())
	valueArr := derefConditionOperands(cond.Values())

	predefinedPredicateFn := this.GetPredefinedPredicate(fieldName, schema.Name())
	if predefinedPredicateFn != nil {
		result, cErr := predefinedPredicateFn(operator, value)
		if len(cErr) > 0 {
			return "", cErr, nil
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
	value = derefConditionOperand(value)
	valueArr = derefConditionOperands(valueArr)

	field, quotedField, err := this.prepareColNameForGraph(ctx, schema, fieldName)
	if err != nil {
		return "", nil, err
	}
	var language *cmodel.LanguageCode
	if ctx != nil {
		language = ctx.language
	}

	switch operator {
	case dmodel.Equals, dmodel.NotEquals, dmodel.GreaterThan,
		dmodel.GreaterEqual, dmodel.LessThan, dmodel.LessEqual:
		return this.comparisonPredicate(sb, quotedField, field, operator, value)
	case dmodel.In, dmodel.NotIn:
		return this.collectionPredicate(sb, quotedField, field, operator, valueArr)
	case dmodel.Contains, dmodel.NotContains, dmodel.StartsWith,
		dmodel.NotStartsWith, dmodel.EndsWith, dmodel.NotEndsWith:
		return this.stringPredicate(sb, quotedField, field, operator, value, language)
	case dmodel.IsSet, dmodel.IsNotSet:
		return nullPredicate(sb, quotedField, operator), nil, nil
	default:
		return "", ClientErrorsUnsupportedFilterOperator(originalField), nil
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
) (string, ft.ClientErrors, error) {
	converted, cErrs, err := this.convertValue(field, value)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	switch op {
	case dmodel.Equals:
		return sb.Equal(quotedField, converted), nil, nil
	case dmodel.NotEquals:
		return sb.NotEqual(quotedField, converted), nil, nil
	case dmodel.GreaterThan:
		return sb.GreaterThan(quotedField, converted), nil, nil
	case dmodel.GreaterEqual:
		return sb.GreaterEqualThan(quotedField, converted), nil, nil
	case dmodel.LessThan:
		return sb.LessThan(quotedField, converted), nil, nil
	case dmodel.LessEqual:
		return sb.LessEqualThan(quotedField, converted), nil, nil
	default:
		panic("comparisonPredicate: unsupported operator (internal)")
	}
}

func (this *PgQueryBuilder) collectionPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *dmodel.ModelField, op dmodel.Operator, values []any,
) (string, ft.ClientErrors, error) {
	converted, cErrs, err := this.convertValues(field, values)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	if op == dmodel.In {
		return sb.In(quotedField, converted...), nil, nil
	}
	if op == dmodel.NotIn {
		return sb.NotIn(quotedField, converted...), nil, nil
	}
	return "", nil, errors.Errorf("collectionPredicate: unsupported collection operator '%s'", op)
}

func (this *PgQueryBuilder) stringPredicate(
	sb *sqlbuilder.SelectBuilder, quotedField string,
	field *dmodel.ModelField, op dmodel.Operator, value any, language *cmodel.LanguageCode,
) (string, ft.ClientErrors, error) {
	expr := quotedField
	if isLangJsonField(field) {
		expr = langJsonStringPredicateExpr(quotedField, language)
	} else if columnCategoryFor(field.ColumnType()) != columnString {
		return "", nil, errors.Errorf(
			"stringPredicate: operator '%s' requires string field '%s'", op, field.Name())
	}
	converted, cErrs, err := this.convertStringPredicateValue(field, value)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	pattern := stringPattern(fmt.Sprint(converted), op)
	if pattern == "" {
		return "", nil, errors.Errorf("stringPredicate: unsupported string operator '%s'", op)
	}
	switch op {
	case dmodel.NotContains, dmodel.NotStartsWith, dmodel.NotEndsWith:
		return sb.NotILike(expr, pattern), nil, nil
	default:
		return sb.ILike(expr, pattern), nil, nil
	}
}

func isLangJsonField(field *dmodel.ModelField) bool {
	return field != nil && field.ColumnType() == "nikkiLangJson"
}

func langJsonStringPredicateExpr(sqlRef string, language *cmodel.LanguageCode) string {
	if language == nil || strings.TrimSpace(string(*language)) == "" {
		return fmt.Sprintf("(%s)::text", sqlRef)
	}
	return fmt.Sprintf("(%s ->> %s)", sqlRef, pgStringLiteral(string(*language)))
}

func (this *PgQueryBuilder) convertStringPredicateValue(field *dmodel.ModelField, value any) (
	any, ft.ClientErrors, error,
) {
	if !isLangJsonField(field) {
		return this.convertValue(field, value)
	}
	if value == nil {
		return nil, ft.ClientErrors{*dmodel.NewInvalidDataTypeErr(field.Name())}, nil
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok || v.Kind() != reflect.String {
		return nil, ft.ClientErrors{*dmodel.NewInvalidDataTypeErr(field.Name())}, nil
	}
	return v.String(), nil, nil
}

func nullPredicate(sb *sqlbuilder.SelectBuilder, quotedField string, op dmodel.Operator) string {
	if op == dmodel.IsSet {
		return sb.IsNotNull(quotedField)
	}
	return sb.IsNull(quotedField)
}

func (this *PgQueryBuilder) convertValues(field *dmodel.ModelField, values []any) ([]any, ft.ClientErrors, error) {
	if len(values) == 0 {
		return nil, nil, errors.Errorf(
			"convertValues: operator requires at least one value for field '%s'", field.Name())
	}
	converted := make([]any, len(values))
	for i, value := range values {
		next, cErrs, err := this.convertValue(field, value)
		if err != nil {
			return nil, nil, err
		}
		if len(cErrs) > 0 {
			return nil, cErrs, nil
		}
		converted[i] = next
	}
	return converted, nil, nil
}

func (this *PgQueryBuilder) orderExprs(
	ctx *graphSelectCtx, schema *dmodel.ModelSchema, order dmodel.SearchOrder,
) ([]string, error) {
	exprs := make([]string, 0, len(order))
	for _, item := range order {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		fieldName := item[0]
		field, ref, err := this.resolveOrderField(ctx, schema, fieldName)
		if err != nil {
			return nil, err
		}
		if field.IsVirtualModelField() {
			return nil, errors.Errorf(
				"orderExprs: order field '%s' is not stored in this schema", fieldName)
		}
		dir := "ASC"
		if item.Direction() == dmodel.Desc {
			dir = "DESC"
		}
		exprs = append(exprs, fmt.Sprintf("%s %s", ref, dir))
	}
	return exprs, nil
}

func (this *PgQueryBuilder) resolveOrderField(
	ctx *graphSelectCtx, schema *dmodel.ModelSchema, fieldName string,
) (*dmodel.ModelField, string, error) {
	if ctx == nil || ctx.planner == nil {
		return this.prepareColName(schema, fieldName)
	}
	return ctx.planner.resolveFieldSqlRef(fieldName, MaxOrderGraphFieldDots)
}

func (this *PgQueryBuilder) convertValue(field *dmodel.ModelField, value any) (any, ft.ClientErrors, error) {
	if field.IsVirtualModelField() {
		return nil, clientErrorsVirtualFieldUnavailable(field.Name()), nil
	}
	if field.IsArray() {
		return convertArrayFieldValue(field, value)
	}
	if value == nil {
		if field.IsNullable() {
			return nil, nil, nil
		}
		return nil, nil, errors.Errorf("convertValue: field '%s' does not allow NULL", field.Name())
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if field.IsNullable() {
			return nil, nil, nil
		}
		return nil, nil, errors.Errorf("convertValue: field '%s' does not allow NULL", field.Name())
	}
	if !valueAllowed(columnCategoryFor(field.ColumnType()), v) {
		return nil, ft.ClientErrors{*dmodel.NewInvalidDataTypeErr(field.Name())}, nil
	}

	if columnCategoryFor(field.ColumnType()) == columnJSON {
		raw, err := json.Marshal(v.Interface())
		if err != nil {
			return nil, ft.ClientErrors{*dmodel.NewInvalidDataTypeErr(field.Name())}, errors.Wrapf(err, "convertValue: field '%s': marshal json", field.Name())
		}

		return string(raw), nil, nil
	}

	return v.Interface(), nil, nil
}

func convertArrayFieldValue(field *dmodel.ModelField, value any) (any, ft.ClientErrors, error) {
	if value == nil {
		if field.IsNullable() {
			return nil, nil, nil
		}
		return nil, nil, errors.Errorf("convertArrayFieldValue: field '%s' does not allow NULL", field.Name())
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		if field.IsNullable() {
			return nil, nil, nil
		}
		return nil, nil, errors.Errorf("convertArrayFieldValue: field '%s' does not allow NULL", field.Name())
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, ft.ClientErrors{*dmodel.NewInvalidDataTypeErr(field.Name())}, nil
	}
	cat := columnCategoryFor(field.ColumnType())
	raw, err := buildPgArrayRaw(field, cat, v)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "convertArrayFieldValue: field '%s'", field.Name())
	}
	return raw, nil, nil
}

func buildPgArrayRaw(field *dmodel.ModelField, cat columnCategory, v reflect.Value) (any, error) {
	castSuffix, err := pgArrayTypeCast(field)
	if err != nil {
		return nil, err
	}
	n := v.Len()
	var b strings.Builder
	b.WriteString("ARRAY[")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		elem, elemErr := pgArrayElementSQL(cat, v.Index(i))
		if elemErr != nil {
			return nil, elemErr
		}
		b.WriteString(elem)
	}
	b.WriteString("]::")
	b.WriteString(castSuffix)
	return sqlbuilder.Raw(b.String()), nil
}

func pgArrayTypeCast(field *dmodel.ModelField) (string, error) {
	base, err := resolveGenericToPgType(field.ColumnType())
	if err != nil {
		return "", err
	}
	return base + "[]", nil
}

func pgArrayElementSQL(cat columnCategory, elem reflect.Value) (string, error) {
	ev, ok := unwrapValue(elem)
	if !ok {
		return "NULL", nil
	}
	switch cat {
	case columnString:
		if ev.Kind() != reflect.String {
			return "", errors.Errorf("pgArrayElementSQL: expected string element, got %s", ev.Kind())
		}
		return "'" + strings.ReplaceAll(ev.String(), "'", "''") + "'", nil
	case columnBool:
		if ev.Kind() != reflect.Bool {
			return "", errors.Errorf("pgArrayElementSQL: expected bool element, got %s", ev.Kind())
		}
		if ev.Bool() {
			return "TRUE", nil
		}
		return "FALSE", nil
	case columnInt:
		if !(isIntKind(ev.Kind()) || isUintKind(ev.Kind())) {
			return "", errors.Errorf("pgArrayElementSQL: expected integer element, got %s", ev.Kind())
		}
		return fmt.Sprintf("%d", ev.Interface()), nil
	case columnNumeric:
		switch ev.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fmt.Sprintf("%d", ev.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return fmt.Sprintf("%d", ev.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(ev.Float(), 'g', -1, 64), nil
		default:
			if isBigNumber(ev.Interface()) {
				return fmt.Sprintf("%v", ev.Interface()), nil
			}
			return "", errors.Errorf("pgArrayElementSQL: unsupported numeric element %s", ev.Kind())
		}
	case columnTime:
		t, tErr := modelTimeFromReflect(ev)
		if tErr != nil {
			return "", tErr
		}
		return "'" + t.UTC().Format("2006-01-02 15:04:05.999999 MST") + "'", nil
	case columnJSON:
		raw, jsonErr := json.Marshal(ev.Interface())
		if jsonErr != nil {
			return "", errors.Errorf("pgArrayElementSQL: cannot marshal json element: %v", jsonErr)
		}
		escaped := strings.ReplaceAll(string(raw), "'", "''")
		return "'" + escaped + "'::jsonb", nil
	default:
		return "", errors.Errorf("pgArrayElementSQL: unsupported column category %v", cat)
	}
}

func modelTimeFromReflect(v reflect.Value) (time.Time, error) {
	x := v.Interface()
	switch t := x.(type) {
	case time.Time:
		return t, nil
	case *time.Time:
		if t == nil {
			return time.Time{}, errors.New("modelTimeFromReflect: nil time pointer")
		}
		return *t, nil
	case cmodel.ModelDateTime:
		return t.GoTime(), nil
	case *cmodel.ModelDateTime:
		if t == nil {
			return time.Time{}, errors.New("modelTimeFromReflect: nil ModelDateTime")
		}
		return t.GoTime(), nil
	case cmodel.ModelDate:
		return t.GoTime(), nil
	case *cmodel.ModelDate:
		if t == nil {
			return time.Time{}, errors.New("modelTimeFromReflect: nil ModelDate")
		}
		return t.GoTime(), nil
	case cmodel.ModelTime:
		return t.GoTime(), nil
	case *cmodel.ModelTime:
		if t == nil {
			return time.Time{}, errors.New("modelTimeFromReflect: nil ModelTime")
		}
		return t.GoTime(), nil
	default:
		return time.Time{}, errors.Errorf("modelTimeFromReflect: unsupported type %T", x)
	}
}

func resolveModelFieldToPgType(col *dmodel.ModelField) (string, error) {
	if col.IsVirtualModelField() {
		return "", errors.Errorf("resolveModelFieldToPgType: virtual field '%s' has no SQL type", col.Name())
	}
	base, err := resolveGenericToPgType(col.ColumnType())
	if err != nil {
		return "", err
	}
	if col.IsArray() {
		return base + "[]", nil
	}
	return base, nil
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
	case dmodel.FieldDataTypeNameEmail, dmodel.FieldDataTypeNamePhone, dmodel.FieldDataTypeNameString,
		dmodel.FieldDataTypeNameSecret, dmodel.FieldDataTypeNameUrl, dmodel.FieldDataTypeNameEnumString,
		dmodel.FieldDataTypeNameUlid, dmodel.FieldDataTypeNameEtag, dmodel.FieldDataTypeNameLangCode,
		"nikkiModelId", dmodel.FieldDataTypeNameSlug:
		return "character varying", nil
	case dmodel.FieldDataTypeNameUuid:
		return "uuid", nil
	case dmodel.FieldDataTypeNameInt64:
		return "bigint", nil
	case "int", "int8", "int16", dmodel.FieldDataTypeNameInt32, "integer", "enumNumber", "enumInteger",
		dmodel.FieldDataTypeNameEnumInt32:
		return "integer", nil
	case dmodel.FieldDataTypeNameDecimal, "float":
		return "numeric", nil
	case dmodel.FieldDataTypeNameBoolean:
		return "boolean", nil
	case "date", dmodel.FieldDataTypeNameModelDate:
		return "date", nil
	case "time", dmodel.FieldDataTypeNameModelTime:
		return "time without time zone", nil
	case "dateTime", dmodel.FieldDataTypeNameModelDateTime:
		return "timestamptz", nil
	case dmodel.FieldDataTypeNameLangJson, dmodel.FieldDataTypeNameJsonMap:
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

func pgStringLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func prependTenantKey(tenantKey string, fields []string) []string {
	if tenantKey == "" {
		return fields
	}
	if slices.Contains(fields, tenantKey) {
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

func buildExistsCaseSqlWithOptCond(tableRef string, columns []string, argIdx int, optCond string) string {
	conds := make([]string, len(columns))
	for i, col := range columns {
		conds[i] = fmt.Sprintf("%s = $%d", pgQuote(col), argIdx+i)
	}

	whereClause := strings.Join(conds, " AND ")
	if optCond != "" {
		whereClause = fmt.Sprintf("%s AND %s", whereClause, optCond)
	}

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

func derefConditionOperand(value any) any {
	if value == nil {
		return nil
	}
	v, ok := unwrapValue(reflect.ValueOf(value))
	if !ok {
		return nil
	}
	return v.Interface()
}

func derefConditionOperands(values []any) []any {
	if len(values) == 0 {
		return values
	}
	out := make([]any, len(values))
	for i := range values {
		out[i] = derefConditionOperand(values[i])
	}
	return out
}

func unwrapValue(v reflect.Value) (reflect.Value, bool) {
	for {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		case reflect.Interface:
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		default:
			return v, true
		}
	}
}

var timeType = reflect.TypeOf(time.Time{})
var modelDateType = reflect.TypeOf(cmodel.NewModelDate())
var modelDateTimeType = reflect.TypeOf(cmodel.NewModelDateTime())
var modelTimeType = reflect.TypeOf(cmodel.NewModelTime())

func isShopspringDecimalValue(v reflect.Value) bool {
	x := v.Interface()
	switch typed := x.(type) {
	case decimal.Decimal:
		return true
	case *decimal.Decimal:
		return typed != nil
	default:
		return false
	}
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
		return isShopspringDecimalValue(v) ||
			isIntKind(v.Kind()) || isUintKind(v.Kind()) ||
			isFloatKind(v.Kind()) || isBigNumber(v.Interface())
	case columnTime:
		return v.Type() == timeType ||
			v.Type() == modelDateType ||
			v.Type() == modelDateTimeType ||
			v.Type() == modelTimeType
	case columnJSON:
		k := v.Kind()
		return k == reflect.Map || k == reflect.Slice || k == reflect.Array
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

func clientErrorsVirtualFieldUnavailable(field string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_virtual_field_unavailable"),
			"field is not available for this operation"),
	}
}

func clientErrsPtr(c ft.ClientErrors) *ft.ClientErrors {
	if len(c) == 0 {
		return nil
	}
	p := c
	return &p
}

func stringSqlGraphOutcome(sql string, graphCErrs ft.ClientErrors, err error) (*string, *ft.ClientErrors, error) {
	if err != nil {
		return stringSqlOutcome("", err)
	}
	if len(graphCErrs) > 0 {
		return nil, clientErrsPtr(graphCErrs), nil
	}
	return &sql, nil, nil
}

func stringSqlOutcome(sql string, err error) (*string, *ft.ClientErrors, error) {
	if err == nil {
		return &sql, nil, nil
	}
	ce, ierr := extractClientErr(err)
	if ierr != nil {
		return nil, nil, ierr
	}
	return nil, ce, nil
}

func extractClientErr(err error) (*ft.ClientErrors, error) {
	if err == nil {
		return nil, nil
	}
	var unk *errClientUnknownField
	if stdErr.As(err, &unk) {
		e := clientErrorsUnknownField(unk.Field)
		return &e, nil
	}
	var sqlErr *errClientSqlErrors
	if stdErr.As(err, &sqlErr) {
		return &sqlErr.errors, nil
	}
	return nil, err
}
