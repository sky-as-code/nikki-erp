package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)


// The JSON model files are parsed at start-up by RegisterModels; a malformed file panics the
// whole app. These tests turn that into a test failure instead, and pin the field set the
// UoM business rules depend on.

func TestUomCatSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := UomCatSchemaBuilder().Build()

	assert.Equal(t, UomCatSchemaName, schema.Name())
	assert.Equal(t, "essential_uomcats", schema.TableName())
	// Without a record label field the frontend relation picker shows raw ULIDs.
	assert.Equal(t, UomCatFieldName, schema.RecordLabelField())
	requireField(t, schema, UomCatFieldReferenceUomId)
}

func TestUomSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := UomSchemaBuilder().Build()

	assert.Equal(t, UomSchemaName, schema.Name())
	assert.Equal(t, "essential_uoms", schema.TableName())
	assert.Equal(t, UomFieldName, schema.RecordLabelField())
	// BR-UOM-ESS-002: a UoM cannot exist outside a category.
	assert.True(t, requireField(t, schema, UomFieldCategoryId).IsRequiredForCreate())
}

// BR-UOM-ESS-008 and BR-UOM-ESS-018 require enough precision to hold factors such as
// 0.453592 without floating-point drift, so both must be decimals rather than integers.
func TestUomFactorAndRoundingAreDecimal(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := UomSchemaBuilder().Build()

	for _, fieldName := range []string{UomFieldFactor, UomFieldRounding} {
		field := requireField(t, schema, fieldName)
		assert.Equal(t, dmodel.FieldDataTypeNameDecimal, field.DataType().String(),
			"field %q must be a decimal", fieldName)
	}
}

// BR-UOM-ESS-009: the three UoM types are a closed set the validation rules switch on.
func TestUomTypeEnumValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, UomSchemaBuilder().Build(), UomFieldUomType)

	assert.ElementsMatch(t,
		[]string{UomTypeReference.String(), UomTypeBiggerEqual.String(), UomTypeSmaller.String()},
		field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues])
}

func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up.
	_ = basemodel.RegisterJsonBaseSchemas()
}

func requireField(t *testing.T, schema *dmodel.ModelSchema, fieldName string) *dmodel.ModelField {
	t.Helper()
	field, ok := schema.Fields()[fieldName]
	require.Truef(t, ok, "schema %q has no field %q", schema.Name(), fieldName)
	return field
}
