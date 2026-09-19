package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// A reservation is transactional data with a lifecycle: it is consumed or released, never
// archived. Expiry is a condition of the clock, so the stored status must not be able to say it.
func TestStockReservationLifecycleFields(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := StockReservationSchemaBuilder().Build()

	_, archivable := schema.Fields()[basemodel.FieldIsArchived]
	assert.False(t, archivable, "a reservation is never archived")

	stored := enumValuesOf(t, requireField(t, schema, StockReservationFieldStatus))
	assert.ElementsMatch(t, []string{"active", "consumed", "released"}, stored)

	effective := enumValuesOf(t, requireField(t, schema, StockReservationFieldEffectiveStatus))
	assert.ElementsMatch(t, []string{"active", "consumed", "released", "expired"}, effective)

	for _, name := range []string{
		StockReservationFieldRemainingQuantity,
		StockReservationFieldEffectiveStatus,
		StockReservationFieldEffectiveReservedQuantity,
	} {
		field := requireField(t, schema, name)
		assert.Truef(t, field.IsVirtual(), "%s is derived, never stored", name)
	}
}

// The source key is opaque and must accept whatever identifier the calling module uses, exactly
// as the transfer's source_id does for the legacy holds.
func TestStockReservationAcceptsOpaqueSourceKeys(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := StockReservationSchemaBuilder().Build()
	field := requireField(t, schema, StockReservationFieldSourceId)

	for _, key := range []string{"01M2JTCJ2NQHHDNYQ97N5JXWQQ", "SO-101/L1", "fulfillment:01K5VMSLOTLOC0000000000000"} {
		value, err := field.Validate(key, true)
		require.Nil(t, err)
		require.Equal(t, key, *value.Get())
	}
}

func TestInventoryEventTypesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, eventType := range InventoryEventTypes() {
		assert.Falsef(t, seen[eventType], "event type %q listed twice", eventType)
		seen[eventType] = true
	}
	assert.Len(t, seen, 6)
}
