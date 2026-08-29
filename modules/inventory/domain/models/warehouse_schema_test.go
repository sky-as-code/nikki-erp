package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Warehouse and Inventory Location are the only two resources with an operational state; the rest
// express their whole lifecycle through archiving. That is why suspend and resume exist on these
// two and nowhere else, and encoding it here makes adding a status elsewhere fail a test.
func TestOnlyWarehouseAndLocationCarryStatus(t *testing.T) {
	requireBaseSchemasRegistered(t)

	withStatus := map[string]func() *dmodel.ModelSchemaBuilder{
		WarehouseSchemaName:         WarehouseSchemaBuilder,
		InventoryLocationSchemaName: InventoryLocationSchemaBuilder,
	}
	withoutStatus := map[string]func() *dmodel.ModelSchemaBuilder{
		StorageCategorySchemaName:         StorageCategorySchemaBuilder,
		WarehouseSupplyRelationSchemaName: WarehouseSupplyRelationSchemaBuilder,
		PutawayRuleSchemaName:             PutawayRuleSchemaBuilder,
	}

	for name, builder := range withStatus {
		_, ok := builder().Build().Fields()["status"]
		assert.Truef(t, ok, "%s needs a status: it can be suspended and resumed", name)
	}
	for name, builder := range withoutStatus {
		_, ok := builder().Build().Fields()["status"]
		assert.Falsef(t, ok, "%s has no state independent of archiving", name)
	}
}

// The status values are active and suspended; 'inactive' and 'blocked' are folded into suspension.
// A stray value would let a record reach a state no operation can move it out of.
func TestWarehouseAndLocationStatusValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	for name, builder := range map[string]func() *dmodel.ModelSchemaBuilder{
		WarehouseSchemaName:         WarehouseSchemaBuilder,
		InventoryLocationSchemaName: InventoryLocationSchemaBuilder,
	} {
		values := enumValuesOf(t, requireField(t, builder().Build(), "status"))

		assert.ElementsMatchf(t, []string{"active", "suspended"}, values,
			"%s must offer exactly active and suspended", name)
	}
}

// enumValuesOf reads an enum field's allowed values from its data-type options, which hold them
// under a well-known key rather than behind an accessor.
func enumValuesOf(t *testing.T, field *dmodel.ModelField) []string {
	t.Helper()

	raw, ok := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues]
	require.Truef(t, ok, "field %q declares no enum values", field.Name())

	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, item.(string))
		}
		return values
	default:
		require.Failf(t, "unexpected enum values", "field %q: %T", field.Name(), raw)
		return nil
	}
}

// Vendor, customer, inventory-loss and shared transit locations belong to no warehouse, so the
// column must stay nullable for them to exist at all.
func TestLocationWarehouseIdIsNullable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := InventoryLocationSchemaBuilder().Build()
	field := requireField(t, schema, InventoryLocationFieldWarehouseId)

	assert.False(t, field.IsRequiredForCreate(),
		"a vendor or shared transit location belongs to no warehouse")
}

// 'scrap' and 'inventory_loss' are different destinations: an adjustment balances against loss, a
// write-off moves to scrap, and the scrap lifecycle rejects a location of the wrong usage so usable
// stock is not counted twice.
func TestLocationUsageValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := InventoryLocationSchemaBuilder().Build()
	values := enumValuesOf(t, requireField(t, schema, InventoryLocationFieldLocationUsage))

	assert.ElementsMatch(t, []string{
		"internal", "customer", "vendor", "inventory_loss", "scrap", "transit", "virtual",
	}, values)
	assert.NotContains(t, values, "supplier", "renamed to vendor by the change request")
}

// The three flow settings, on both directions. A fourth value would have no topology behind it:
// location provisioning and the movement plan are both switches over exactly these.
func TestWarehouseFlowValues(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := WarehouseSchemaBuilder().Build()
	for _, fieldName := range []string{WarehouseFieldIncomingFlow, WarehouseFieldOutgoingFlow} {
		values := enumValuesOf(t, requireField(t, schema, fieldName))
		assert.ElementsMatchf(t,
			[]string{"one_step", "two_step", "three_step"}, values, "field %q", fieldName)
	}
}
