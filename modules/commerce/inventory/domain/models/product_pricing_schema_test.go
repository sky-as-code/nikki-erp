package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Building a schema proves the JSON parses, not that a computed field can be answered; that needs
// FinalizeRelations over the whole registered module. The registry is global and un-resettable, so
// only TestProductSchemasRegisterInOrder may do it and the resolution assertions live there.

// The template owns the one base sales price; a variant must not copy it. Deriving
// effective_base_sales_price is what makes a template price change move every variant at once,
// which a stored copy would quietly break.
func TestBaseSalesPriceLivesOnTemplateOnly(t *testing.T) {
	requireBaseSchemasRegistered(t)

	template := ProductTemplateSchemaBuilder().Build()
	variant := ProductVariantSchemaBuilder().Build()

	_, onTemplate := template.Fields()[ProductTemplateFieldBaseSalesPrice]
	assert.True(t, onTemplate, "template must own base_sales_price")

	_, onVariant := variant.Fields()[ProductTemplateFieldBaseSalesPrice]
	assert.False(t, onVariant,
		"variant must not store its own base_sales_price; it derives effective_base_sales_price")
}

// Cost lives on the variant: two variants of one product genuinely cost different amounts and
// nothing relates them. The template exposes it only as a read-only proxy, never a stored column.
func TestCostLivesOnVariantAndIsNotStoredOnTemplate(t *testing.T) {
	requireBaseSchemasRegistered(t)

	variant := ProductVariantSchemaBuilder().Build()
	costField, onVariant := variant.Fields()[ProductVariantFieldCost]
	require.True(t, onVariant, "variant must own cost")
	assert.False(t, costField.IsComputed(), "variant cost is stored input, not derived")

	template := ProductTemplateSchemaBuilder().Build()
	if templateCost, present := template.Fields()["cost"]; present {
		assert.True(t, templateCost.IsComputed(),
			"a cost on the template may only be a computed proxy of its single variant's cost "+
				"(BR-PRICE-VARIANT-011), never a second stored value")
	}
}

// Both halves of the sum must be reachable and the surcharge must be the template-scoped one: the
// global inventory_product_attribute_value.price_extra is deprecated because the same value
// surcharges differently on different products.
func TestSalesPriceExtraIsTemplateScoped(t *testing.T) {
	requireBaseSchemasRegistered(t)

	templateValue := ProductTemplateAttributeValueSchemaBuilder().Build()
	_, present := templateValue.Fields()[ProductTemplateAttributeValueFieldSalesPriceExtra]
	assert.True(t, present,
		"the authoritative surcharge belongs to Template x Attribute x Value")

	variantValue := ProductVariantAttributeValueSchemaBuilder().Build()
	copied, present := variantValue.Fields()[ProductVariantAttributeValueFieldSalesPriceExtra]
	require.True(t, present, "the variant junction must carry the denormalised copy")
	assert.False(t, copied.IsComputed(),
		"the copy must be a real column: an aggregate cannot read a computed field, which is the "+
			"whole reason it is denormalised rather than related")
}

// There is no generic Product Price resource, and no second vaguer price field may grow back on
// the product. The forbidden names are the shapes actually reached for. purchase_price matters
// most: a product holds no authoritative purchase price at all, because a vendor's price is
// qualified by vendor, quantity and validity, which one column cannot express.
func TestNoGenericOrPurchasePriceOnProduct(t *testing.T) {
	requireBaseSchemasRegistered(t)

	forbidden := []string{
		"price", "list_price", "proposed_price", "sale_price",
		"purchase_price", "default_purchase_price", "standard_price",
	}

	for _, schema := range []struct {
		name   string
		fields map[string]*dmodel.ModelField
	}{
		{"template", ProductTemplateSchemaBuilder().Build().Fields()},
		{"variant", ProductVariantSchemaBuilder().Build().Fields()},
	} {
		for _, fieldName := range forbidden {
			_, present := schema.fields[fieldName]
			assert.Falsef(t, present,
				"%s must not carry %q: selling price is base_sales_price, cost is the variant's, "+
					"and a vendor price belongs to Purchase", schema.name, fieldName)
		}
	}
}

