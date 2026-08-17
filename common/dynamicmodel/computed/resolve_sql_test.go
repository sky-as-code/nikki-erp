package computed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Fixtures: an order with a one:many "lines" edge (FK on the line) and its line schema. The
// computed fields under test are injected on the order side.

func sqlLineSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cf_sql_line").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("order_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("quantity").DataType(dmodel.FieldDataTypeInt32(0, 100000))).
		Field(dmodel.DefineField().Name("unit_price").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2))).
		Field(dmodel.DefineField().Name("status").DataType(dmodel.FieldDataTypeString(0, 20))).
		EdgeTo(dmodel.Edge("order").ManyToOne("cf_sql_order", dmodel.DynamicFields{"order_id": "id"})).
		Build()
}

func sqlOrderSchema(computedFields ...*dmodel.FieldBuilder) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cf_sql_order").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("customer").DataType(dmodel.FieldDataTypeString(0, 100))).
		EdgeFrom(dmodel.Edge("lines").Existing("cf_sql_line", "order"))
	for _, field := range computedFields {
		builder.Field(field)
	}
	return builder.Build()
}

func finalizeSqlOrder(t *testing.T, computedFields ...*dmodel.FieldBuilder) (*dmodel.SchemaRegistry, error) {
	t.Helper()
	reg := newRegistryWith(t, sqlLineSchema(), sqlOrderSchema(computedFields...))
	return reg, reg.FinalizeRelations()
}

func TestFinalize_AggregateCountPlan(t *testing.T) {
	_, err := finalizeSqlOrder(t,
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)))
	require.NoError(t, err)

	plan := computed.PlanFor("cf_sql_order")
	require.NotNil(t, plan)
	fieldPlan := plan.Fields["line_count"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, computed.TypeInt64, fieldPlan.Type)
	require.NotNil(t, fieldPlan.SqlSource)
	assert.Equal(t, "lines", fieldPlan.SqlSource.Edge)
	assert.Equal(t, "cf_sql_line", fieldPlan.SqlSource.SourceSchemaName)
	assert.False(t, fieldPlan.SqlSource.Many)
	assert.Equal(t, "order_id", fieldPlan.SqlSource.SourceFkColumn)
	assert.Equal(t, "id", fieldPlan.SqlSource.RootRefColumn)
	assert.Empty(t, fieldPlan.PhysicalOperands, "a subquery correlates in SQL; it needs no projected operands")
}

func TestFinalize_AggregateSumInnerExpressionType(t *testing.T) {
	_, err := finalizeSqlOrder(t,
		dmodel.DefineField().Name("total_amount").DataType(dmodel.FieldDataTypeDecimal("0", "999999999", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum,
				computed.AggExpr(computed.Mul(computed.F("quantity"), computed.F("unit_price"))))))
	require.NoError(t, err)

	fieldPlan := computed.PlanFor("cf_sql_order").Fields["total_amount"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, computed.TypeDecimal, fieldPlan.Type)
	assert.Contains(t, fieldPlan.Dependencies, computed.FieldRef{Schema: "cf_sql_line", Field: "quantity"})
	assert.Contains(t, fieldPlan.Dependencies, computed.FieldRef{Schema: "cf_sql_line", Field: "unit_price"})
}

func TestFinalize_ExistsAndLookupPlans(t *testing.T) {
	_, err := finalizeSqlOrder(t,
		dmodel.DefineField().Name("has_open_line").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "open"))),
		dmodel.DefineField().Name("last_price").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Lookup("lines", "unit_price", computed.Desc("id"))))
	require.NoError(t, err)

	plan := computed.PlanFor("cf_sql_order")
	assert.Equal(t, computed.TypeBoolean, plan.Fields["has_open_line"].Type)
	assert.Equal(t, computed.TypeDecimal, plan.Fields["last_price"].Type)
	assert.Contains(t, plan.Fields["has_open_line"].Dependencies,
		computed.FieldRef{Schema: "cf_sql_line", Field: "status"})
}

func TestFinalize_SqlKindRejections(t *testing.T) {
	cases := map[string]*dmodel.FieldBuilder{
		"to-one edge as source": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeInt64(0, 10)).
			Computed(false, computed.Aggregate("order_self", computed.AggCount)),
		"unknown edge": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeInt64(0, 10)).
			Computed(false, computed.Aggregate("nonexistent", computed.AggCount)),
		"dotted source": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeInt64(0, 10)).
			Computed(false, computed.Aggregate("lines.order", computed.AggCount)),
		"unknown operand field": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum, computed.AggField("no_such"))),
		"sum of non-numeric": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum, computed.AggField("status"))),
		"comparison in inner expression": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum,
				computed.AggExpr(computed.Gt(computed.F("quantity"), computed.Lit(int64(0)))))),
		"go function in inner expression": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum,
				computed.AggExpr(computed.Fn("round", computed.F("unit_price"))))),
		"filter on unknown field": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("no_such", dmodel.Equals, "x"))),
		"filter with linked operator": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Linked, "x"))),
		"undeclared context key": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, computed.Ctx("company_id")))),
		"declared but unused context key": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "open"), "company_id")),
		"foreign placeholder value": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines",
				dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "${id}"))),
		"lookup order_by unknown field": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Lookup("lines", "unit_price", computed.Desc("no_such"))),
		"lookup dotted order_by": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Lookup("lines", "unit_price", computed.Desc("order.id"))),
		"declared type mismatch": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeString(0, 20)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)),
		"default type mismatch": dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeInt64(0, 10)).
			Computed(false, computed.Aggregate("lines", computed.AggCount, computed.AggDefault("zero"))),
	}
	for name, field := range cases {
		t.Run(name, func(t *testing.T) {
			schemas := []*dmodel.ModelSchema{sqlLineSchema(), sqlOrderSchemaWithSelfEdge(field)}
			reg := newRegistryWith(t, schemas...)
			err := reg.FinalizeRelations()
			require.Error(t, err)
		})
	}
}

// sqlOrderSchemaWithSelfEdge adds a to-one edge so the "to-one edge as source" case has a target.
func sqlOrderSchemaWithSelfEdge(computedFields ...*dmodel.FieldBuilder) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cf_sql_order").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("customer").DataType(dmodel.FieldDataTypeString(0, 100))).
		Field(dmodel.DefineField().Name("parent_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeFrom(dmodel.Edge("lines").Existing("cf_sql_line", "order")).
		EdgeTo(dmodel.Edge("order_self").ManyToOne("cf_sql_order", dmodel.DynamicFields{"parent_id": "id"}))
	for _, field := range computedFields {
		builder.Field(field)
	}
	return builder.Build()
}

func TestFinalize_FilterDepthLimit(t *testing.T) {
	leaf := dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "open")
	nested := leaf
	for i := 0; i < 6; i++ {
		nested = dmodel.NewSearchNode().And(*nested)
	}
	_, err := finalizeSqlOrder(t,
		dmodel.DefineField().Name("deep").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Exists("lines", nested)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeding the maximum")
}

func TestFinalize_AggregateOfComputedSourceFieldRejected(t *testing.T) {
	line := dmodel.DefineModel("cf_sql_line").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("order_id").DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().Name("quantity").DataType(dmodel.FieldDataTypeInt32(0, 100000))).
		Field(dmodel.DefineField().Name("double_qty").DataType(dmodel.FieldDataTypeInt64(0, 200000)).
			Computed(false, computed.Mul(computed.F("quantity"), computed.Lit(int64(2))))).
		EdgeTo(dmodel.Edge("order").ManyToOne("cf_sql_order", dmodel.DynamicFields{"order_id": "id"})).
		Build()
	order := sqlOrderSchema(
		dmodel.DefineField().Name("bad").DataType(dmodel.FieldDataTypeDecimal("0", "9", 2)).
			Computed(false, computed.Aggregate("lines", computed.AggSum, computed.AggField("double_qty"))))
	reg := newRegistryWith(t, line, order)

	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reference computed field")
}

func TestFinalize_GoExpressionMayDependOnSqlComputedField(t *testing.T) {
	_, err := finalizeSqlOrder(t,
		dmodel.DefineField().Name("line_count").DataType(dmodel.FieldDataTypeInt64(0, 1000000)).
			Computed(false, computed.Aggregate("lines", computed.AggCount)),
		dmodel.DefineField().Name("is_busy").DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Gt(computed.F("line_count"), computed.Lit(int64(10)))))
	require.NoError(t, err)

	plan := computed.PlanFor("cf_sql_order")
	busy := plan.Fields["is_busy"]
	require.NotNil(t, busy)
	assert.Equal(t, []string{"line_count"}, busy.ComputedDeps)
	assert.Empty(t, busy.PhysicalOperands, "the SQL-computed dependency arrives via its subquery, not the projection")
}
