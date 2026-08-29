package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// engineBackedSchemas lists the engine-served schemas in the registration order their edges
// require. Mirrors InventoryModule.RegisterModels, which cannot be imported here without a cycle.
func engineBackedSchemas(t *testing.T) []*dmodel.ModelSchema {
	t.Helper()
	requireBaseSchemasRegistered(t)

	// Built, not registered: the registry is process-wide and TestProductSchemasRegisterInOrder
	// owns registration, so registering here too would collide.
	builders := []*dmodel.ModelSchemaBuilder{
		ProductTypeSchemaBuilder(), ProductCategorySchemaBuilder(), BrandSchemaBuilder(),
		ProductAttributeSchemaBuilder(), ProductAttributeValueSchemaBuilder(),
		ProductTemplateSchemaBuilder(), ProductTemplateAttributeSchemaBuilder(),
		ProductTemplateAttributeValueSchemaBuilder(),
		ProductVariantSchemaBuilder(), ProductVariantAttributeValueSchemaBuilder(),
		WarehouseSchemaBuilder(), StorageCategorySchemaBuilder(),
		InventoryLocationSchemaBuilder(),
		WarehouseSupplyRelationSchemaBuilder(), PutawayRuleSchemaBuilder(),
		StockOperationTypeSchemaBuilder(), StockQuantSchemaBuilder(),
		StockTransferSchemaBuilder(), StockMoveSchemaBuilder(), StockMoveLineSchemaBuilder(),
		StockMoveDependencySchemaBuilder(),
		StockScrapSchemaBuilder(), StockProductConfigSchemaBuilder(),
	}
	schemas := make([]*dmodel.ModelSchema, 0, len(builders))
	for _, builder := range builders {
		schemas = append(schemas, builder.Build())
	}
	return schemas
}

// An omitted default_search_fields list is legal to the builder, so a schema missing it silently
// degrades to a primary-keys-only listing. This test turns that back into a failure.
func TestEverySchemaDeclaresDefaultSearchFields(t *testing.T) {
	for _, schema := range engineBackedSchemas(t) {
		assert.NotEmptyf(t, schema.DefaultSearchFields(),
			"schema %q lists no default_search_fields, so its listing returns only primary keys",
			schema.Name())
	}
}

// knownOpaqueDefaultFields are schemas with no human-readable field to show instead; replacing
// them needs an edge that does not exist yet.
var knownOpaqueDefaultFields = map[string]bool{
	// Junction rows whose every column is a foreign key; dropping them leaves no columns at all.
	"inventory_product_variant_attribute_value.product_variant_id":          true,
	"inventory_product_variant_attribute_value.template_attribute_value_id": true,
	"inventory_stock_move_dependency.predecessor_move_id":                   true,
	"inventory_stock_move_dependency.successor_move_id":                     true,
	// The config row is identified by the template it configures; it owns no name.
	"inventory_stock_product_config.product_template_id": true,
}

// These fields become the listing's columns, so a foreign key or edge shows a raw ULID where a
// name belongs.
func TestDefaultSearchFieldsAreDisplayable(t *testing.T) {
	for _, schema := range engineBackedSchemas(t) {
		for _, fieldName := range schema.DefaultSearchFields() {
			field, ok := schema.Fields()[fieldName]
			if !assert.Truef(t, ok, "schema %q: default field %q does not exist", schema.Name(), fieldName) {
				continue
			}
			if knownOpaqueDefaultFields[schema.Name()+"."+fieldName] {
				continue
			}
			assert.Falsef(t, field.IsEdgeModel(),
				"schema %q: default field %q is an edge", schema.Name(), fieldName)
			assert.Falsef(t, field.IsForeignKey(),
				"schema %q: default field %q is a foreign key, which renders as a raw ULID",
				schema.Name(), fieldName)
		}
	}
}
