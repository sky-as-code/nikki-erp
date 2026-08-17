package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// SQL-string assertions for the correlated-subquery emitter — no live database, following the
// house style of the other pg_query_builder tests.

func computedOrderLineSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cf_ormc_line").
		ShouldBuildDb().
		TableName("cf_ormc_lines").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("order_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("quantity").DataType(dmodel.FieldDataTypeInt32(0, 100000))).
		Field(dmodel.DefineField().Name("unit_price").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2))).
		Field(dmodel.DefineField().Name("status").DataType(dmodel.FieldDataTypeString(0, 20))).
		Field(dmodel.DefineField().Name("is_archived").DataType(dmodel.FieldDataTypeBoolean())).
		EdgeTo(dmodel.Edge("order").ManyToOne("cf_ormc_order", dmodel.DynamicFields{"order_id": "id"})).
		Build()
}

func computedOrderSchema(computedFields ...*dmodel.FieldBuilder) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cf_ormc_order").
		ShouldBuildDb().
		TableName("cf_ormc_orders").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		EdgeFrom(dmodel.Edge("lines").Existing("cf_ormc_line", "order"))
	for _, field := range computedFields {
		builder.Field(field)
	}
	return builder.Build()
}

func computedEmitFixture(t *testing.T, fieldName string, field *dmodel.FieldBuilder) (
	*dmodel.SchemaRegistry, *dmodel.ModelSchema, *computed.FieldPlan,
) {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(computedOrderLineSchema()))
	order := computedOrderSchema(field)
	require.NoError(t, reg.Register(order))
	require.NoError(t, reg.FinalizeRelations())
	plan := computed.PlanFor("cf_ormc_order")
	require.NotNil(t, plan)
	fieldPlan := plan.Fields[fieldName]
	require.NotNil(t, fieldPlan)
	return reg, order, fieldPlan
}

func emitComputedSql(t *testing.T, fieldName string, field *dmodel.FieldBuilder, ctxValues map[string]any) string {
	t.Helper()
	reg, order, fieldPlan := computedEmitFixture(t, fieldName, field)
	builder := NewPgQueryBuilder().(*PgQueryBuilder)
	sql, cErrs, err := builder.ComputedSubqueryExpr(reg, order, "t1", fieldPlan, ctxValues)
	require.NoError(t, err)
	require.Empty(t, cErrs)
	return sql
}

func TestComputedSubquery_AggregateCountWithFilterAndDefault(t *testing.T) {
	sql := emitComputedSql(t, "completed_count",
		dmodel.DefineField().Name("completed_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount,
				computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "completed")),
				computed.AggDefault(int64(0)))), nil)

	assert.Contains(t, sql, "COALESCE((SELECT COUNT(*) FROM")
	assert.Contains(t, sql, `"cf_ormc_lines"`)
	assert.Contains(t, sql, `"order_id" = t1."id"`, "the subquery must correlate to the aliased root row")
	assert.Contains(t, sql, `"is_archived" = FALSE`, "archived source rows never count")
	assert.Contains(t, sql, "completed")
	assert.Contains(t, sql, ", 0)", "the declared default wraps the subquery")
	assert.NotContains(t, sql, "GROUP BY", "aggregation is correlated, never JOIN+GROUP BY")
	assert.NotContains(t, sql, "JOIN")
}

func TestComputedSubquery_SumInnerExpression(t *testing.T) {
	sql := emitComputedSql(t, "total_amount",
		dmodel.DefineField().Name("total_amount").DataType(dmodel.FieldDataTypeDecimal("0", "999999999", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum,
				computed.AggExpr(computed.Mul(computed.F("quantity"), computed.F("unit_price"))))), nil)

	assert.Contains(t, sql, `SUM(("quantity" * "unit_price"))`)
}

func TestComputedSubquery_DivisionCastsToNumeric(t *testing.T) {
	sql := emitComputedSql(t, "half_qty",
		dmodel.DefineField().Name("half_qty").DataType(dmodel.FieldDataTypeDecimal("0", "999999", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum,
				computed.AggExpr(computed.Div(computed.F("quantity"), computed.Lit(int64(2)))))), nil)

	assert.Contains(t, sql, `("quantity")::numeric /`,
		"SQL division must match the Go evaluator's always-decimal semantics")
}

func TestComputedSubquery_ExistsForm(t *testing.T) {
	sql := emitComputedSql(t, "has_open",
		dmodel.DefineField().Name("has_open").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "open"))), nil)

	assert.Contains(t, sql, "EXISTS (SELECT 1 FROM")
	assert.NotContains(t, sql, "COUNT", "existence compiles to EXISTS, not COUNT(*) > 0")
}

func TestComputedSubquery_LookupOrderedLimitOne(t *testing.T) {
	sql := emitComputedSql(t, "last_price",
		dmodel.DefineField().Name("last_price").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Lookup("lines", "unit_price",
				computed.Desc("id"), computed.Asc("status"))), nil)

	assert.Contains(t, sql, `(SELECT "unit_price" FROM`)
	assert.Contains(t, sql, `ORDER BY "id" DESC, "status" ASC`)
	assert.Contains(t, sql, "LIMIT 1")
}

func TestComputedSubquery_ContextSubstitution(t *testing.T) {
	field := dmodel.DefineField().Name("ctx_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
		Computed(false, computed.Aggregate("lines", computed.AggCount,
			computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, computed.Ctx("wanted_status"))),
			computed.AggContext("wanted_status")))

	sql := emitComputedSql(t, "ctx_count", field, map[string]any{"wanted_status": "done"})
	assert.Contains(t, sql, "done", "the bound context value must reach the SQL as a literal")
	assert.NotContains(t, sql, "${ctx.", "no placeholder may survive into SQL")
}

func TestComputedSubquery_MissingContextValueIsClientError(t *testing.T) {
	field := dmodel.DefineField().Name("ctx_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
		Computed(false, computed.Aggregate("lines", computed.AggCount,
			computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, computed.Ctx("wanted_status"))),
			computed.AggContext("wanted_status")))
	reg, order, fieldPlan := computedEmitFixture(t, "ctx_count", field)

	builder := NewPgQueryBuilder().(*PgQueryBuilder)
	_, cErrs, err := builder.ComputedSubqueryExpr(reg, order, "t1", fieldPlan, nil)
	require.NoError(t, err)
	require.Len(t, cErrs, 1)
	assert.Equal(t, ft.ErrorKey("err_computed_context_missing"), cErrs[0].Key)
}

func TestComputedSubquery_RequiresRootAlias(t *testing.T) {
	reg, order, fieldPlan := computedEmitFixture(t, "line_count",
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))

	builder := NewPgQueryBuilder().(*PgQueryBuilder)
	_, _, err := builder.ComputedSubqueryExpr(reg, order, "", fieldPlan, nil)
	require.Error(t, err)
}

// End-to-end projection: SqlSelectGraph must materialize a requested SQL-computed field as an
// aliased correlated subquery in the SELECT, with the root aliased for correlation.

func TestSelectGraph_ProjectsSqlComputedFieldAsSubquery(t *testing.T) {
	reg, order, _ := computedEmitFixture(t, "line_count",
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))

	builder := NewPgQueryBuilder()
	sql, cErrs, err := builder.SqlSelectGraph(order, reg, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id", "line_count"}),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)

	assert.Contains(t, *sql, `(SELECT COUNT(*) FROM "cf_ormc_lines" WHERE "order_id" = t0."id"`)
	assert.Contains(t, *sql, `AS "line_count"`)
	assert.Contains(t, *sql, `FROM "cf_ormc_orders" AS t0`, "the root must be aliased for correlation")
	assert.NotContains(t, *sql, "GROUP BY")
}

func TestSelectGraph_OnlySqlComputedFieldStillAnchorsOnPk(t *testing.T) {
	reg, order, _ := computedEmitFixture(t, "line_count",
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))

	builder := NewPgQueryBuilder()
	sql, cErrs, err := builder.SqlSelectGraph(order, reg, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"line_count"}),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.Contains(t, *sql, `AS "line_count"`)
}

func TestSelectGraph_WildcardDoesNotProjectSqlComputedFields(t *testing.T) {
	reg, order, _ := computedEmitFixture(t, "line_count",
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))

	builder := NewPgQueryBuilder()
	sql, cErrs, err := builder.SqlSelectGraph(order, reg, nil, SqlSelectGraphOpts{})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)
	assert.NotContains(t, *sql, "SELECT COUNT", "SQL-computed fields are opt-in by explicit projection")
}

func TestSelectGraph_SqlComputedPerRequestLimit(t *testing.T) {
	restore := computed.ActiveLimits()
	limits := restore
	limits.MaxSqlComputedFieldsPerRequest = 1
	computed.SetLimits(limits)
	defer computed.SetLimits(restore)

	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(computedOrderLineSchema()))
	order := computedOrderSchema(
		dmodel.DefineField().Name("count_a").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)),
		dmodel.DefineField().Name("count_b").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))
	require.NoError(t, reg.Register(order))
	require.NoError(t, reg.FinalizeRelations())

	builder := NewPgQueryBuilder()
	_, cErrs, err := builder.SqlSelectGraph(order, reg, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"count_a", "count_b"}),
	})
	if err != nil {
		assert.Contains(t, err.Error(), "err_too_many_sql_computed_fields")
		return
	}
	require.NotNil(t, cErrs)
	require.Greater(t, cErrs.Count(), 0)
}

func TestSelectGraph_TwoAggregatesStayFanOutSafe(t *testing.T) {
	reg := dmodel.NewSchemaRegistry()
	require.NoError(t, reg.Register(computedOrderLineSchema()))
	order := computedOrderSchema(
		dmodel.DefineField().Name("count_all").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)),
		dmodel.DefineField().Name("total_qty").DataType(dmodel.FieldDataTypeDecimal("0", "999999999", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum, computed.AggField("quantity"))))
	require.NoError(t, reg.Register(order))
	require.NoError(t, reg.FinalizeRelations())

	builder := NewPgQueryBuilder()
	sql, cErrs, err := builder.SqlSelectGraph(order, reg, nil, SqlSelectGraphOpts{
		Columns: ToSelectColumns([]string{"id", "count_all", "total_qty"}),
	})
	require.NoError(t, err)
	require.Nil(t, cErrs)
	require.NotNil(t, sql)

	// AC-06: both aggregates are independent correlated subqueries — the root query never joins
	// the collection, so one aggregate can never inflate the other.
	assert.Contains(t, *sql, "COUNT(*)")
	assert.Contains(t, *sql, `SUM("quantity")`)
	assert.NotContains(t, *sql, "JOIN")
	assert.NotContains(t, *sql, "GROUP BY")
}

// Injection: the value seams of the SQL-computed kinds. Identifiers are attacked in
// computed/injection_test.go (they die at finalize); the values below DO reach SQL and must
// survive only as escaped literals.

const computedSqlPayload = `x'; DROP TABLE cf_ormc_lines; --`

func assertEscapedOnce(t *testing.T, sql string) {
	t.Helper()
	assert.Contains(t, sql, `E'x\'; DROP TABLE cf_ormc_lines; --'`,
		"the payload must survive only as an escaped string literal")
	assert.NotContains(t, sql, `'x'; DROP`,
		"an unescaped quote would terminate the literal and execute the rest")
	assert.Equal(t, 1, strings.Count(sql, "DROP TABLE"), "the payload appears once, inside its literal")
}

func TestComputedInjection_HostileFilterValueIsEscaped(t *testing.T) {
	sql := emitComputedSql(t, "evil_count",
		dmodel.DefineField().Name("evil_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount,
				computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, computedSqlPayload)))), nil)

	assertEscapedOnce(t, sql)
}

func TestComputedInjection_HostileContextValueIsEscaped(t *testing.T) {
	field := dmodel.DefineField().Name("evil_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
		Computed(false, computed.Aggregate("lines", computed.AggCount,
			computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, computed.Ctx("status"))),
			computed.AggContext("status")))

	sql := emitComputedSql(t, "evil_count", field, map[string]any{"status": computedSqlPayload})
	assertEscapedOnce(t, sql)
}

func TestComputedInjection_HostileDefaultIsEscaped(t *testing.T) {
	sql := emitComputedSql(t, "evil_lookup",
		dmodel.DefineField().Name("evil_lookup").DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Lookup("lines", "status",
				computed.Desc("id"), computed.LookupDefault(computedSqlPayload))), nil)

	// The default renders through pgStringLiteral, which doubles embedded quotes — the other
	// safe form: the payload's quote stays inside the literal instead of terminating it.
	assert.Contains(t, sql, "COALESCE(")
	assert.Contains(t, sql, `'x''; DROP TABLE cf_ormc_lines; --'`,
		"the payload must survive only as an escaped string literal")
	assert.NotContains(t, sql, `'x'; DROP`,
		"an unescaped quote would terminate the literal and execute the rest")
	assert.Equal(t, 1, strings.Count(sql, "DROP TABLE"), "the payload appears once, inside its literal")
}

func TestComputedInjection_HostileInValuesAreEscaped(t *testing.T) {
	sql := emitComputedSql(t, "evil_count",
		dmodel.DefineField().Name("evil_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount,
				computed.AggFilter(dmodel.NewSearchNode().NewCondition("status", dmodel.In, "ok", computedSqlPayload)))), nil)

	assertEscapedOnce(t, sql)
}

func TestComputedSubquery_LookupWithoutDefaultHasNoCoalesce(t *testing.T) {
	sql := emitComputedSql(t, "last_status",
		dmodel.DefineField().Name("last_status").DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Lookup("lines", "status", computed.Desc("id"))), nil)

	assert.NotContains(t, sql, "COALESCE", "no default declared: a missing record must read as NULL")
}
