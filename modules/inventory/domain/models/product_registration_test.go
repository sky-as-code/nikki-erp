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
	}

	for _, builder := range builders {
		schema := builder.Build()
		require.NoErrorf(t, dmodel.RegisterSchemaB(builder), "failed to register %q", schema.Name())
	}
}
