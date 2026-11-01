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
	requireEssentialSchemasRegistered(t)

	builders := []*dmodel.ModelSchemaBuilder{
		ProductCategorySchemaBuilder(),
		BrandSchemaBuilder(),
		ProductAttributeSchemaBuilder(),
		ProductAttributeValueSchemaBuilder(),
		ProductTemplateSchemaBuilder(),
		ProductTemplateAttributeSchemaBuilder(),
		ProductTemplateAttributeValueSchemaBuilder(),
		ProductTemplateCombinationExclusionSchemaBuilder(),
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
	assertTemplateCostAndUnitsResolved(t)
}

// assertTemplateCostAndUnitsResolved checks the template's cost and unit fields. The template holds
// a stored cost, the starting cost of a variant created without one; the variant-cost range it
// used to expose is gone. The unit names are function computed fields resolved from org settings.
func assertTemplateCostAndUnitsResolved(t *testing.T) {
	t.Helper()

	template := dmodel.GetSchemaRegistry().Get(ProductTemplateSchemaName)
	require.NotNil(t, template)

	cost, hasCost := template.Fields()[ProductTemplateFieldCost]
	require.True(t, hasCost, "the template must hold a cost")
	require.False(t, cost.IsComputed(), "the template cost is stored, not derived from its variants")

	for _, removed := range []string{"min_variant_cost", "max_variant_cost"} {
		_, present := template.Fields()[removed]
		require.Falsef(t, present, "%q was removed with the template cost", removed)
	}
	for _, name := range []string{ProductTemplateFieldWeightUomSymbol, ProductTemplateFieldDimensionUomSymbol} {
		field, present := template.Fields()[name]
		require.Truef(t, present, "template must expose %q", name)
		require.Truef(t, field.IsComputed(), "%q is resolved from settings, never stored", name)
	}
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
