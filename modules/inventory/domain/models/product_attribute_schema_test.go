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

// BR §4.7: the three creation modes are a closed set that variant generation switches on. A
// fourth value appearing here without generation logic to match would silently create no
// variants at all.
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

// BR §6.5.3 and §14.3 step 2: only instant and dynamic attributes contribute to variant
// identity. This pins the rule the combination-key builder depends on.
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
	// BR §6.6.1: a value cannot exist outside an attribute.
	assert.True(t, requireField(t, schema, ProductAttributeValueFieldAttributeId).IsRequiredForCreate())
}

// BR §6.6.1: price_extra is money, so it must be a decimal rather than a float that would drift.
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
