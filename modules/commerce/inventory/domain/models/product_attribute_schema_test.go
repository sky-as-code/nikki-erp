package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestProductAttributeSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductAttributeSchemaBuilder().Build()

	assert.Equal(t, ProductAttributeSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_attributes", schema.TableName())
	assert.Equal(t, ProductAttributeFieldName, schema.RecordLabelField())
}

// The three creation modes are a closed set variant generation switches on; a fourth value with
// no matching generation logic would silently create no variants.
func TestVariantCreationModeEnumValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, ProductAttributeSchemaBuilder().Build(), ProductAttributeFieldVariantCreationMode)

	assert.ElementsMatch(t,
		[]string{
			VariantCreationModeInstant.String(),
			VariantCreationModeDynamic.String(),
			VariantCreationModeNever.String(),
		},
		field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues])
}

// Only instant and dynamic attributes contribute to variant identity; the combination-key builder
// depends on this.
func TestVariantCreationModeCreatesVariants(t *testing.T) {
	assert.True(t, VariantCreationModeInstant.CreatesVariants())
	assert.True(t, VariantCreationModeDynamic.CreatesVariants())
	assert.False(t, VariantCreationModeNever.CreatesVariants(),
		"a NEVER attribute must be excluded from the combination key")
}

func TestAttributeDataTypeEnumValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, ProductAttributeSchemaBuilder().Build(), ProductAttributeFieldDataType)

	assert.ElementsMatch(t,
		[]string{
			AttributeDataTypeOption.String(),
			AttributeDataTypeText.String(),
			AttributeDataTypeNumber.String(),
			AttributeDataTypeDate.String(),
			AttributeDataTypeBoolean.String(),
		},
		field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues])
}

func TestProductAttributeValueSchemaParses(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := ProductAttributeValueSchemaBuilder().Build()

	assert.Equal(t, ProductAttributeValueSchemaName, schema.Name())
	assert.Equal(t, "inventory_product_attribute_values", schema.TableName())
	assert.Equal(t, ProductAttributeValueFieldName, schema.RecordLabelField())
	// A value cannot exist outside an attribute.
	assert.True(t, requireField(t, schema, ProductAttributeValueFieldAttributeId).IsRequiredForCreate())
}

// price_extra is money, so it must be a decimal rather than a float that would drift.
func TestProductAttributeValuePriceExtraIsDecimal(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field := requireField(t, ProductAttributeValueSchemaBuilder().Build(), ProductAttributeValueFieldPriceExtra)

	assert.Equal(t, dmodel.FieldDataTypeNameDecimal, field.DataType().String())
}

func TestAttributeModelsAreArchivable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	builders := map[string]*dmodel.ModelSchemaBuilder{
		ProductAttributeSchemaName:      ProductAttributeSchemaBuilder(),
		ProductAttributeValueSchemaName: ProductAttributeValueSchemaBuilder(),
	}

	for name, builder := range builders {
		schema := builder.Build()
		_, ok := schema.Fields()[basemodel.FieldIsArchived]
		assert.Truef(t, ok, "schema %q must extend core.basemodel.archivable_model", name)
	}
}
