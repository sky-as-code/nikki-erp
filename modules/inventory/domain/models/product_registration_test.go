package models

import (
	"testing"

	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Registering every schema at once mirrors start-up and is the only place cross-schema edges are
// resolved: a schema registered before the one its edge points at fails here rather than panicking
// on boot. The order below is InventoryModule.RegisterModels' order; keep the two in step.
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

	// FinalizeRelations resolves edges and compiles computed fields, rejecting one that cannot be
	// answered: an aggregate crossing more than one edge, or over an operand that is itself computed
	// and so has no column. It runs here because the registry is global with no reset, so a second
	// test registering this module would fail on the first schema.
	require.NoError(t, dmodel.GetSchemaRegistry().FinalizeRelations(),
		"every registered schema must finalize, including the variant's derived pricing fields")

	assertVariantPricingFieldsResolved(t)
	assertTemplateCostRangeResolved(t)
}

// assertTemplateCostRangeResolved checks the template's cost read model. The template has no cost
// column: cost belongs to the variant, since two variants of one product genuinely cost different
// amounts. It exposes the range across its variants instead, and that range must never be
// collapsed into a single number called the product's cost.
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

// effective_base_sales_price is an expression over a related field plus an aggregate over a
// collection edge. Replacing that with a two-hop source or an aggregate over a computed field
// would fail at boot rather than here.
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
