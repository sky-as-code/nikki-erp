package computed_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func finalizeVariantFixture(t *testing.T) {
	t.Helper()
	reg := newRegistryWith(t, templateLikeSchema(), variantLikeSchema(
		dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name")),
	))
	require.NoError(t, reg.FinalizeRelations())
}

func finalizeChainFixture(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	schema := dmodel.DefineModel("cf_eval_chain").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt32(0, 100000))).
		Field(dmodel.DefineField().Name("price").DataType(dmodel.FieldDataTypeDecimal("0", "99999", 2))).
		Field(dmodel.DefineField().Name("subtotal").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Mul(computed.F("qty"), computed.F("price")))).
		Field(dmodel.DefineField().Name("total").DataType(dmodel.FieldDataTypeDecimal("0", "9999999", 2)).
			Computed(false, computed.Mul(computed.F("subtotal"), computed.Lit(int64(2))))).
		Build()
	reg := newRegistryWith(t, schema)
	require.NoError(t, reg.FinalizeRelations())
	return schema
}

func TestBuildEvalPlan_ExplicitProjection(t *testing.T) {
	finalizeVariantFixture(t)

	plan, errs := computed.BuildEvalPlan("cf_fin_variant", []string{"id", "template_name"})
	require.Equal(t, 0, errs.Count())
	require.NotNil(t, plan)

	assert.Equal(t, []string{"template_name"}, plan.Wanted)
	assert.Equal(t, []string{"template_id"}, plan.ExtraFields,
		"the FK join key must be appended to an explicit projection")
	require.Len(t, plan.RelatedReads, 1)
	read := plan.RelatedReads[0]
	assert.Equal(t, "cf_fin_template", read.SchemaName)
	assert.Equal(t, "template_id", read.FkColumn)
	assert.Equal(t, "id", read.RefColumn)
	assert.Equal(t, map[string]string{"template_name": "name"}, read.Leaves)
}

func TestBuildEvalPlan_NoComputedWanted(t *testing.T) {
	finalizeVariantFixture(t)

	plan, errs := computed.BuildEvalPlan("cf_fin_variant", []string{"id", "template_id"})
	require.Equal(t, 0, errs.Count())
	assert.Nil(t, plan, "a projection naming no computed field must pass through untouched")

	plan, errs = computed.BuildEvalPlan("cf_unknown_schema", nil)
	require.Equal(t, 0, errs.Count())
	assert.Nil(t, plan)
}

func TestBuildEvalPlan_EmptyProjectionWantsAll(t *testing.T) {
	finalizeVariantFixture(t)

	plan, errs := computed.BuildEvalPlan("cf_fin_variant", nil)
	require.Equal(t, 0, errs.Count())
	require.NotNil(t, plan)
	assert.Equal(t, []string{"template_name"}, plan.Wanted)
	assert.Empty(t, plan.ExtraFields, "no explicit projection means every column comes back anyway")
}

func TestBuildEvalPlan_DependencyClosure(t *testing.T) {
	finalizeChainFixture(t)

	plan, errs := computed.BuildEvalPlan("cf_eval_chain", []string{"id", "total"})
	require.Equal(t, 0, errs.Count())
	require.NotNil(t, plan)

	assert.Equal(t, []string{"subtotal", "total"}, plan.Wanted,
		"an unrequested dependency must still evaluate, before its dependent")
	assert.ElementsMatch(t, []string{"qty", "price"}, plan.ExtraFields)
}

func TestBuildEvalPlan_PerRequestLimit(t *testing.T) {
	limits := computed.DefaultLimits()
	limits.MaxComputedFieldsPerRequest = 1
	computed.SetLimits(limits)
	defer computed.SetLimits(computed.DefaultLimits())
	finalizeChainFixture(t)

	_, errs := computed.BuildEvalPlan("cf_eval_chain", []string{"total"})
	assert.Greater(t, errs.Count(), 0, "total plus its dependency exceeds the limit of 1")
}

func TestEvalPlanApply_RelatedBatchedFill(t *testing.T) {
	finalizeVariantFixture(t)
	plan, _ := computed.BuildEvalPlan("cf_fin_variant", []string{"id", "template_name"})
	require.NotNil(t, plan)

	rows := []dmodel.DynamicFields{
		{"id": "v1", "template_id": "t1"},
		{"id": "v2", "template_id": "t2"},
		{"id": "v3", "template_id": "t1"},
		{"id": "v4", "template_id": nil},
		{"id": "v5", "template_id": "t9"}, // dead reference
	}

	var gotSchema, gotKeyColumn string
	var gotKeys []any
	var gotFields []string
	calls := 0
	search := func(schemaName string, keyColumn string, keys []any, fields []string) ([]dmodel.DynamicFields, error) {
		calls++
		gotSchema, gotKeyColumn, gotKeys, gotFields = schemaName, keyColumn, keys, fields
		return []dmodel.DynamicFields{
			{"id": "t1", "name": "Widget"},
			{"id": "t2", "name": "Gadget"},
		}, nil
	}

	require.NoError(t, plan.Apply(rows, search))

	assert.Equal(t, 1, calls, "one batched read per edge per page, never one per row")
	assert.Equal(t, "cf_fin_template", gotSchema)
	assert.Equal(t, "id", gotKeyColumn)
	assert.ElementsMatch(t, []any{"t1", "t2", "t9"}, gotKeys, "keys are distinct, nil skipped")
	assert.ElementsMatch(t, []string{"id", "name"}, gotFields)

	assert.Equal(t, "Widget", rows[0]["template_name"])
	assert.Equal(t, "Gadget", rows[1]["template_name"])
	assert.Equal(t, "Widget", rows[2]["template_name"])
	_, present := rows[3]["template_name"]
	assert.False(t, present, "a row without a source key keeps the field absent")
	_, present = rows[4]["template_name"]
	assert.False(t, present, "a dead reference reads as unknown, not as empty values")
}

func TestEvalPlanApply_ExpressionChain(t *testing.T) {
	finalizeChainFixture(t)
	plan, _ := computed.BuildEvalPlan("cf_eval_chain", []string{"id", "total"})
	require.NotNil(t, plan)

	rows := []dmodel.DynamicFields{
		{"id": "r1", "qty": int64(3), "price": decimal.NewFromInt(10)},
		{"id": "r2", "qty": nil, "price": decimal.NewFromInt(10)},
	}
	require.NoError(t, plan.Apply(rows, nil))

	assert.True(t, decimal.NewFromInt(30).Equal(rows[0]["subtotal"].(decimal.Decimal)))
	assert.True(t, decimal.NewFromInt(60).Equal(rows[0]["total"].(decimal.Decimal)))
	assert.Nil(t, rows[1]["subtotal"], "null operand propagates")
	assert.Nil(t, rows[1]["total"])
}

func TestRejectWrites(t *testing.T) {
	schema := buildQuantLikeSchema(t)

	errs := computed.RejectWrites(schema, dmodel.DynamicFields{
		"on_hand_quantity":   "10",
		"available_quantity": "999",
	})
	require.Equal(t, 1, errs.Count())
	assert.Contains(t, errs.ToError().Error(), `Field "available_quantity" is computed and cannot be written`)

	errs = computed.RejectWrites(schema, dmodel.DynamicFields{"on_hand_quantity": "10"})
	assert.Equal(t, 0, errs.Count())
}
