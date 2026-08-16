package computed_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func buildQuantLikeSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	return dmodel.DefineModel("cf_test_quant").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("on_hand_quantity").
			DataType(dmodel.FieldDataTypeDecimal("0", "999999999", 4))).
		Field(dmodel.DefineField().Name("reserved_quantity").
			DataType(dmodel.FieldDataTypeDecimal("0", "999999999", 4))).
		Field(dmodel.DefineField().Name("available_quantity").
			DataType(dmodel.FieldDataTypeDecimal("-999999999", "999999999", 4)).
			Computed(false, computed.Sub(
				computed.Fn("coalesce", computed.F("on_hand_quantity"), computed.Lit(0)),
				computed.Fn("coalesce", computed.F("reserved_quantity"), computed.Lit(0)),
			))).
		Build()
}

func TestComputedField_IsVirtualAndComputed(t *testing.T) {
	schema := buildQuantLikeSchema(t)
	field, ok := schema.Field("available_quantity")
	require.True(t, ok)

	assert.True(t, field.IsComputed())
	assert.True(t, field.IsVirtual(), "a computed field must be virtual: every no-column behavior is inherited")
	assert.False(t, field.IsPersisted())

	plain, ok := schema.Field("on_hand_quantity")
	require.True(t, ok)
	assert.False(t, plain.IsComputed())
}

func TestComputedField_ExcludedFromColumnsIncludedInReadable(t *testing.T) {
	schema := buildQuantLikeSchema(t)

	assert.NotContains(t, fieldNames(schema.Columns()), "available_quantity")
	assert.Contains(t, fieldNames(schema.ReadableFields()), "available_quantity")
}

func fieldNames(fields []*dmodel.ModelField) []string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name())
	}
	return names
}

func TestComputedField_DroppedFromWrites(t *testing.T) {
	schema := buildQuantLikeSchema(t)

	validated, errs := schema.Validate(dmodel.DynamicFields{
		"on_hand_quantity":   "10",
		"reserved_quantity":  "3",
		"available_quantity": "999",
	})
	require.Equal(t, 0, errs.Count(), errs)
	_, present := validated["available_quantity"]
	assert.False(t, present, "the existing virtual-field write-strip must apply to computed fields")
}

func TestComputedField_DefOfDerivesKind(t *testing.T) {
	schema := buildQuantLikeSchema(t)
	field, _ := schema.Field("available_quantity")

	def, err := computed.DefOf(field)
	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, computed.ComputeExpression, def.Kind)
	assert.False(t, def.IsStored)
	assert.NotNil(t, def.Expression)

	plain, _ := schema.Field("on_hand_quantity")
	def, err = computed.DefOf(plain)
	require.NoError(t, err)
	assert.Nil(t, def)
}

func TestComputedField_RelatedKindThroughBuilder(t *testing.T) {
	schema := dmodel.DefineModel("cf_test_variant").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("template_name").
			DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name"))).
		Build()

	field, _ := schema.Field("template_name")
	def, err := computed.DefOf(field)
	require.NoError(t, err)
	assert.Equal(t, computed.ComputeRelated, def.Kind)
	assert.Equal(t, "template.name", def.Related)
}

func TestComputedField_StoredIsRejectedAtBuild(t *testing.T) {
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "Build must panic on is_stored: true")
		assert.Contains(t, panicMessage(recovered), "stored computed fields are not yet supported")
	}()
	dmodel.DefineModel("cf_test_stored").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("total").
			DataType(dmodel.FieldDataTypeDecimal("0", "999", 2)).
			Computed(true, computed.F("subtotal"))).
		Build()
}

func TestComputedField_NilExpressionPanicsImmediately(t *testing.T) {
	assert.Panics(t, func() {
		dmodel.DefineField().Name("broken").
			DataType(dmodel.FieldDataTypeString(0, 10)).
			Computed(false, nil)
	})
}

func TestComputedField_RoleContradictionsPanicAtBuild(t *testing.T) {
	// A computed field is virtual, so the virtual-field contradiction checks must fire for it.
	assert.Panics(t, func() {
		dmodel.DefineModel("cf_test_contradiction").
			Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
			Field(dmodel.DefineField().Name("derived").
				DataType(dmodel.FieldDataTypeString(0, 10)).
				Computed(false, computed.F("id")).
				Unique()).
			Build()
	})
}

func TestComputedField_DefOfRejectsForeignExpressionType(t *testing.T) {
	schema := dmodel.DefineModel("cf_test_foreign").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("odd").
			DataType(dmodel.FieldDataTypeString(0, 10)).
			Computed(false, "not an expression")).
		Build()

	field, _ := schema.Field("odd")
	_, err := computed.DefOf(field)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a computed.Expr")
}

func TestComputedField_ToSimplizedReportsComputedFlags(t *testing.T) {
	schema := buildQuantLikeSchema(t)
	field, _ := schema.Field("available_quantity")

	simplized := field.ToSimplized()
	asMap := toJsonMap(t, simplized)
	assert.Equal(t, true, asMap["is_computed"])
	assert.Equal(t, true, asMap["is_virtual"])
	assert.Equal(t, false, asMap["is_persisted"])
	assert.Equal(t, false, asMap["is_edge_model"])
	// Read-only, but not server-owned: a computed field carries business meaning and must stay
	// available to a client's column picker.
	assert.Equal(t, false, asMap["is_system_field"])
}

func toJsonMap(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	var asMap map[string]any
	require.NoError(t, json.Unmarshal(raw, &asMap))
	return asMap
}

func panicMessage(recovered any) string {
	if err, ok := recovered.(error); ok {
		return err.Error()
	}
	return fmt.Sprint(recovered)
}
