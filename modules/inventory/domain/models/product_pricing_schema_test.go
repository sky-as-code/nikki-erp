package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The pricing fields added by the product-pricing change request, checked at the level that
// actually catches a mistake.
//
// Building a schema proves the JSON parses; it does not prove a computed field can be answered.
// That resolution happens in FinalizeRelations, and asserting it needs the whole module registered
// — which the schema registry is global and un-resettable about, so exactly one test may do it.
// That test is TestProductSchemasRegisterInOrder in product_registration_test.go, and the pricing
// assertion lives there rather than being duplicated here.

// The base sales price is the one price the template owns. A variant must NOT carry a copy of it:
// the whole point of deriving effective_base_sales_price is that raising the template's price
// moves every variant at once (BR-PRICE-VARIANT-004), which a stored copy would quietly break.
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

// Cost is the mirror image: it lives on the VARIANT, because two variants of one product really
// do cost different amounts and nothing relates them (BR-PRICE-VARIANT-006). The template exposes
// it only as a read-only proxy, never as a stored column of its own.
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

// Both halves of the sum must be reachable, and the surcharge must be the template-scoped one.
// The global inventory_product_attribute_value.price_extra is deprecated precisely because the
// same value surcharges differently on different products (BR-PRICE-VARIANT-002).
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

// The generic Product Price resource is gone (PRICE-INV-001, AC-PRICE-044), and nothing may
// quietly grow back in its place.
//
// This replaces an assertion that used to live in product_price_schema_test.go, deleted with the
// resource. Its premise has inverted — the old test said price "lives on inventory_product_price",
// whereas now the template owns base_sales_price outright — but the rule it protected is exactly
// what this change request turns into an invariant: no second, vaguer price field on the product.
//
// The named fields are the shapes that would actually be reached for: a bare `price`, a Odoo-style
// `list_price`, or a `purchase_price` on the product. That last one matters most. Product must
// hold no authoritative purchase price at all (PRICE-INV-011, AC-PRICE-003) — a vendor's price
// belongs to Purchase, where it is qualified by vendor, quantity and validity, none of which a
// single column on a product can express.
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

