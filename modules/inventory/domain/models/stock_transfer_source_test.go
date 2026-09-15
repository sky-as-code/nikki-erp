package models

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// Sales holds each slot separately using fulfillmentID:locationID (Sales fulfilment docs).
// Inventory must preserve that external key, including the suffix used to release the hold.
func TestStockTransferAcceptsReservationSourceKeys(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := StockTransferSchemaBuilder().Build()
	field, ok := schema.Field(StockTransferFieldSourceId)
	require.True(t, ok)
	fulfillment := "01M2JTCJ2NQHHDNYQ97N5JXWQQ"
	for _, key := range []string{fulfillment, fulfillment + ":01K5VMSLOTLOC0000000000000", fulfillment + ":01K5VMSLOTLOC0000000000001"} {
		t.Run(key, func(t *testing.T) {
			value, err := field.Validate(key, true)
			require.Nil(t, err, "external reservation keys must not be constrained to ULIDs")
			require.Equal(t, key, *value.Get())
		})
	}
}
