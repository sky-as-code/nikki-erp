package computed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func newRegistryWith(t *testing.T, schemas ...*dmodel.ModelSchema) *dmodel.SchemaRegistry {
	t.Helper()
	reg := dmodel.NewSchemaRegistry()
	for _, schema := range schemas {
		require.NoError(t, reg.Register(schema))
	}
	return reg
}

func templateLikeSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cf_fin_template").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

func variantLikeSchema(computedFields ...*dmodel.FieldBuilder) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cf_fin_variant").
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("template_id").DataType(dmodel.FieldDataTypeUlid())).
		EdgeTo(dmodel.Edge("template").ManyToOne("cf_fin_template", dmodel.DynamicFields{"template_id": "id"}))
	for _, field := range computedFields {
		builder.Field(field)
	}
	return builder.Build()
}

func TestFinalize_RelatedFieldPlan(t *testing.T) {
	reg := newRegistryWith(t, templateLikeSchema(), variantLikeSchema(
		dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name")),
	))

	require.NoError(t, reg.FinalizeRelations())

	plan := computed.PlanFor("cf_fin_variant")
	require.NotNil(t, plan)
	fieldPlan := plan.Fields["template_name"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, "template", fieldPlan.RelatedEdge)
	assert.Equal(t, "name", fieldPlan.RelatedLeaf)
	assert.Equal(t, "cf_fin_template", fieldPlan.RelatedSchemaName)
	assert.Equal(t, "template_id", fieldPlan.RelatedFkColumn)
	assert.Equal(t, "id", fieldPlan.RelatedRefColumn)
	assert.Equal(t, []string{"template_id"}, fieldPlan.PhysicalOperands)
	assert.Equal(t, computed.Type("string"), fieldPlan.Type)
}

func TestFinalize_ExpressionDependencyChainAndEvalOrder(t *testing.T) {
	schema := dmodel.DefineModel("cf_fin_chain").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt32(0, 100000))).
		Field(dmodel.DefineField().Name("price").DataType(dmodel.FieldDataTypeDecimal("0", "99999", 2))).
		Field(dmodel.DefineField().Name("total").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Mul(computed.F("subtotal"), computed.Lit(int64(2))))).
		Field(dmodel.DefineField().Name("subtotal").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Mul(computed.F("qty"), computed.F("price")))).
		Build()
	reg := newRegistryWith(t, schema)

	require.NoError(t, reg.FinalizeRelations())

	plan := computed.PlanFor("cf_fin_chain")
	require.NotNil(t, plan)

	total := plan.Fields["total"]
	assert.Equal(t, []string{"subtotal"}, total.ComputedDeps)
	assert.ElementsMatch(t, []string{"qty", "price"}, total.PhysicalOperands,
		"the dependency's operands must fold into the dependent's projection needs")

	subtotalIdx, totalIdx := -1, -1
	for i, name := range plan.EvalOrder {
		switch name {
		case "subtotal":
			subtotalIdx = i
		case "total":
			totalIdx = i
		}
	}
	assert.Less(t, subtotalIdx, totalIdx, "a dependency must evaluate before its dependent")
}

func TestFinalize_CycleDetected(t *testing.T) {
	schema := dmodel.DefineModel("cf_fin_cycle").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("a").DataType(dmodel.FieldDataTypeInt32(0, 100)).
			Computed(false, computed.Add(computed.F("b"), computed.Lit(int64(1))))).
		Field(dmodel.DefineField().Name("b").DataType(dmodel.FieldDataTypeInt32(0, 100)).
			Computed(false, computed.Add(computed.F("a"), computed.Lit(int64(1))))).
		Build()
	reg := newRegistryWith(t, schema)

	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Computed field dependency cycle detected:")
	assert.Regexp(t, `cf_fin_cycle\.(a|b) -> cf_fin_cycle\.(a|b) -> cf_fin_cycle\.(a|b)`, err.Error())
}

func TestFinalize_SelfReferenceIsACycle(t *testing.T) {
	schema := dmodel.DefineModel("cf_fin_self").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("a").DataType(dmodel.FieldDataTypeInt32(0, 100)).
			Computed(false, computed.Add(computed.F("a"), computed.Lit(int64(1))))).
		Build()
	reg := newRegistryWith(t, schema)

	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cf_fin_self.a -> cf_fin_self.a")
}

func TestFinalize_InvalidDefinitionsRejected(t *testing.T) {
	cases := map[string]*dmodel.FieldBuilder{
		"unknown operand field": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeInt32(0, 100)).
			Computed(false, computed.F("missing")),
		"unknown relation": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Related("nowhere.name")),
		"unknown leaf": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Related("template.nope")),
		"path too deep": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Related("template.brand.name")),
		"bare field path": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Related("name")),
		"declared type mismatch": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeBoolean()).
			Computed(false, computed.Mul(computed.F("template_id"), computed.Lit(int64(2)))),
		"string times ulid": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Mul(computed.Lit("x"), computed.F("template_id"))),
		"unknown function": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.Fn("execute_sql", computed.Lit("DROP TABLE x"))),
		"extract with non-literal part": dmodel.DefineField().Name("bad").
			DataType(dmodel.FieldDataTypeInt32(0, 10000)).
			Computed(false, computed.Fn("extract", computed.F("template_id"), computed.Fn("today"))),
	}

	for name, field := range cases {
		t.Run(name, func(t *testing.T) {
			reg := newRegistryWith(t, templateLikeSchema(), variantLikeSchema(field))
			err := reg.FinalizeRelations()
			require.Error(t, err)
		})
	}
}

func TestFinalize_ExpressionDepthLimit(t *testing.T) {
	var deep computed.Expr = computed.F("qty")
	for range [12]int{} {
		deep = computed.Add(deep, computed.Lit(int64(1)))
	}
	schema := dmodel.DefineModel("cf_fin_deep").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt32(0, 100))).
		Field(dmodel.DefineField().Name("deep").DataType(dmodel.FieldDataTypeInt32(0, 100)).
			Computed(false, deep)).
		Build()
	reg := newRegistryWith(t, schema)

	err := reg.FinalizeRelations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeding the maximum")
}

func TestFinalize_ImpactAnalysis(t *testing.T) {
	reg := newRegistryWith(t, templateLikeSchema(), variantLikeSchema(
		dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name")),
	))
	require.NoError(t, reg.FinalizeRelations())

	dependents := computed.Dependents("cf_fin_template", "name")
	require.Len(t, dependents, 1)
	assert.Equal(t, "cf_fin_variant.template_name", dependents[0].String())

	err := computed.AssertNoDependents("cf_fin_template", "name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Cannot delete field "name".`)
	assert.Contains(t, err.Error(), "- cf_fin_variant.template_name")

	// The FK columns and the edge itself are dependencies too.
	assert.Error(t, computed.AssertNoDependents("cf_fin_template", "id"))
	assert.Error(t, computed.AssertNoDependents("cf_fin_variant", "template"))
	assert.Error(t, computed.AssertNoDependents("cf_fin_variant", "template_id"))
	// An element no computed field touches is safe to remove.
	assert.NoError(t, computed.AssertNoDependents("cf_fin_variant", "id"))
}
