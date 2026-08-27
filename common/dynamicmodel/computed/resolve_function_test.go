package computed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func functionKindSchema(name string, computedFields ...*dmodel.FieldBuilder) *dmodel.ModelSchema {
	builder := dmodel.DefineModel(name).
		ShouldBuildDb().
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("sales_tax_mode").DataType(dmodel.FieldDataTypeString(0, 20)))
	for _, field := range computedFields {
		builder.Field(field)
	}
	return builder.Build()
}

// The motivating case: a list-valued field whose two candidate sources are chosen between by a
// scalar on the same row. No declarative kind can express it, which is why the function kind
// exists — so the array return type must survive resolution untouched.
func TestFinalize_FunctionFieldPlanCarriesArrayType(t *testing.T) {
	reg := newRegistryWith(t, functionKindSchema("cf_fn_variant",
		dmodel.DefineField().Name("effective_sales_tax_ids").
			DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Computed(false, computed.GoFunction("inventory.effective_sales_tax_ids").
				DependsOn("sales_tax_mode").Build()),
	))

	require.NoError(t, reg.FinalizeRelations())

	plan := computed.PlanFor("cf_fn_variant")
	require.NotNil(t, plan)
	fieldPlan := plan.Fields["effective_sales_tax_ids"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, "inventory.effective_sales_tax_ids", fieldPlan.FunctionName)
	assert.Equal(t, "sales_tax_mode", fieldPlan.DependsOn)
	// The dependency must be projected, or the function would receive rows without the field it
	// branches on.
	assert.Equal(t, []string{"sales_tax_mode"}, fieldPlan.PhysicalOperands)
}

func TestFinalize_FunctionFieldWithoutDependsOn(t *testing.T) {
	reg := newRegistryWith(t, functionKindSchema("cf_fn_nodep",
		dmodel.DefineField().Name("derived").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.GoFunction("some.fn").Build()),
	))

	require.NoError(t, reg.FinalizeRelations())

	fieldPlan := computed.PlanFor("cf_fn_nodep").Fields["derived"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, "some.fn", fieldPlan.FunctionName)
	assert.Empty(t, fieldPlan.DependsOn)
	assert.Empty(t, fieldPlan.PhysicalOperands)
}

func TestFinalize_FunctionRejectsUnknownDependsOn(t *testing.T) {
	reg := newRegistryWith(t, functionKindSchema("cf_fn_baddep",
		dmodel.DefineField().Name("derived").
			DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.GoFunction("some.fn").DependsOn("no_such_field").Build()),
	))

	err := reg.FinalizeRelations()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no_such_field")
}

// A function may depend on another computed field; that edge has to be recorded so cycle
// detection and impact analysis both see it.
func TestFinalize_FunctionDependsOnComputedField(t *testing.T) {
	schema := dmodel.DefineModel("cf_fn_chain").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt32(0, 1000))).
		Field(dmodel.DefineField().Name("doubled").DataType(dmodel.FieldDataTypeInt64(0, 2000)).
			Computed(false, computed.Mul(computed.F("qty"), computed.Lit(int64(2))))).
		Field(dmodel.DefineField().Name("derived").DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.GoFunction("some.fn").DependsOn("doubled").Build())).
		Build()
	reg := newRegistryWith(t, schema)

	require.NoError(t, reg.FinalizeRelations())

	fieldPlan := computed.PlanFor("cf_fn_chain").Fields["derived"]
	require.NotNil(t, fieldPlan)
	assert.Equal(t, []string{"doubled"}, fieldPlan.ComputedDeps)
	// The dependency's own operands fold in, so the projection carries what evaluation needs.
	assert.Equal(t, []string{"qty"}, fieldPlan.PhysicalOperands)
}

// A function field is an ordinary node in the dependency graph, so a loop through one must be
// caught by the same cycle detection that guards the declarative kinds.
func TestFinalize_FunctionParticipatesInCycleDetection(t *testing.T) {
	schema := dmodel.DefineModel("cf_fn_cycle").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("left").DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.GoFunction("some.fn").DependsOn("right").Build())).
		Field(dmodel.DefineField().Name("right").DataType(dmodel.FieldDataTypeString(0, 100)).
			Computed(false, computed.GoFunction("other.fn").DependsOn("left").Build())).
		Build()
	reg := newRegistryWith(t, schema)

	err := reg.FinalizeRelations()

	require.Error(t, err)
}
