package models

import (
	"testing"

	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Registering every Products schema in one go is what the app does at start-up, and it is the
// only place cross-schema edges are actually resolved. A schema registered before the one its
// edge points at fails here rather than panicking the whole app on boot.
//
// The order below is the order InventoryModule.RegisterModels uses; keep the two in step.
func TestProductSchemasRegisterInOrder(t *testing.T) {
	requireBaseSchemasRegistered(t)

	builders := []*dmodel.ModelSchemaBuilder{
		ProductTypeSchemaBuilder(),
		ProductCategorySchemaBuilder(),
		BrandSchemaBuilder(),
		ProductAttributeSchemaBuilder(),
		ProductAttributeValueSchemaBuilder(),
		ProductTemplateSchemaBuilder(),
		ProductTemplateAttributeSchemaBuilder(),
		ProductTemplateAttributeValueSchemaBuilder(),
		ProductVariantSchemaBuilder(),
		ProductVariantAttributeValueSchemaBuilder(),
		WarehouseSchemaBuilder(),
		StorageCategorySchemaBuilder(),
		InventoryLocationSchemaBuilder(),
		WarehouseSupplyRelationSchemaBuilder(),
		PutawayRuleSchemaBuilder(),
		StockOperationTypeSchemaBuilder(),
		StockQuantSchemaBuilder(),
		StockTransferSchemaBuilder(),
		StockMoveSchemaBuilder(),
		StockMoveLineSchemaBuilder(),
		StockMoveDependencySchemaBuilder(),
		StockScrapSchemaBuilder(),
	}

	for _, builder := range builders {
		schema := builder.Build()
		require.NoErrorf(t, dmodel.RegisterSchemaB(builder), "failed to register %q", schema.Name())
	}

	// Registering is only half of what start-up does. FinalizeRelations resolves the edges AND
	// compiles every computed field, and it is the step that rejects a computed field that cannot
	// be answered — an aggregate whose source traverses more than one edge, or whose operand is
	// itself computed and so has no column to query.
	//
	// It runs here rather than in its own test because the schema registry is global and has no
	// reset: a second test registering this module fails on the first schema. So the assertions
	// that need a finalized registry belong to whichever test owns it, which is this one.
	require.NoError(t, dmodel.GetSchemaRegistry().FinalizeRelations(),
		"every registered schema must finalize, including the variant's derived pricing fields")

	assertVariantPricingFieldsResolved(t)
	assertTemplateCostRangeResolved(t)
}

// assertTemplateCostRangeResolved checks the template's cost READ MODEL.
//
// The template deliberately has no cost column: cost belongs to the concrete variant, because two
// variants of one product genuinely cost different amounts (BR-PRICE-VARIANT-006). What it exposes
// instead is the range across its variants, so a caller can tell "one cost" from "several" without
// a second query — and BR-PRICE-VARIANT-014 forbids collapsing that range back into a single
// number and calling it the product's cost.
func assertTemplateCostRangeResolved(t *testing.T) {
	t.Helper()

	template := dmodel.GetSchemaRegistry().Get(ProductTemplateSchemaName)
	require.NotNil(t, template)

	for _, name := range []string{
		ProductTemplateFieldMinVariantCost,
		ProductTemplateFieldMaxVariantCost,
	} {
		field, present := template.Fields()[name]
		require.Truef(t, present, "template must expose %q", name)
		require.Truef(t, field.IsComputed(), "%q is a read model, never a stored value", name)
	}

	_, hasOwnCost := template.Fields()[ProductVariantFieldCost]
	require.False(t, hasOwnCost,
		"the template must NOT hold a cost column: cost is the variant's, and a second stored "+
			"value here would be an authoritative product cost that BR-PRICE-VARIANT-006 denies exists")
}

// assertVariantPricingFieldsResolved checks the pricing read model added by the product-pricing
// change request. effective_base_sales_price is an expression over a related field and an
// aggregate over a collection edge; if that composition were ever replaced by a two-hop source or
// an aggregate over a computed field, it would fail at boot rather than here.
func assertVariantPricingFieldsResolved(t *testing.T) {
	t.Helper()

	variant := dmodel.GetSchemaRegistry().Get(ProductVariantSchemaName)
	require.NotNil(t, variant)

	for _, name := range []string{
		ProductVariantFieldTemplateBaseSalesPrice,
		ProductVariantFieldSalesPriceExtraTotal,
		ProductVariantFieldEffectiveBaseSalesPrice,
	} {
		field, present := variant.Fields()[name]
		require.Truef(t, present, "variant must expose %q", name)
		require.Truef(t, field.IsComputed(), "%q must be derived, never stored", name)
	}
}
