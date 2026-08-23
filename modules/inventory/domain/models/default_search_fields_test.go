package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// engineBackedSchemas are the schemas Inventory serves through the dynamic resource engine, in
// the registration order their edges require. It mirrors InventoryModule.RegisterModels; the
// dynamicengines package cannot be imported here without a cycle.
func engineBackedSchemas(t *testing.T) []*dmodel.ModelSchema {
	t.Helper()
	requireBaseSchemasRegistered(t)

	// Built, not registered: TestProductSchemasRegisterInOrder owns registration, and the
	// registry is process-wide, so registering here too would collide with it. Build() resolves
	// everything these assertions read.
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

// The listing's columns come from default_search_fields. The schema builder only rejects names
// that do not resolve to a field -- an omitted list is legal -- so a schema missed while moving
// these out of the Go engine specs degrades to primary-keys-only with no error anywhere. This
// test is what turns that back into a failure.
func TestEverySchemaDeclaresDefaultSearchFields(t *testing.T) {
	for _, schema := range engineBackedSchemas(t) {
		assert.NotEmptyf(t, schema.DefaultSearchFields(),
			"schema %q lists no default_search_fields, so its listing returns only primary keys",
			schema.Name())
	}
}

// knownOpaqueDefaultFields are schemas with no human-readable field to show instead. Replacing
// these needs a computed related field (a peer's name), which needs an edge that does not exist
// yet. Listing them keeps the rule enforced everywhere else rather than switching the check off.
var knownOpaqueDefaultFields = map[string]bool{
	// Junction rows whose every column is a foreign key. Dropping them would leave the listing
	// with no columns at all, which is worse than a raw id.
	"inventory_product_variant_attribute_value.product_variant_id":          true,
	"inventory_product_variant_attribute_value.template_attribute_value_id": true,
	"inventory_stock_move_dependency.predecessor_move_id":                   true,
	"inventory_stock_move_dependency.successor_move_id":                     true,
	// The config row is identified by the template it configures; it owns no name of its own.
	"inventory_stock_product_config.product_template_id": true,
}

// These fields become the listing's columns, so a foreign key or an edge shows the reader a raw
// ULID where a name belongs.
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
