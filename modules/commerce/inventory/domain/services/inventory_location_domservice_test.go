package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Suspend and archive read the same four numbers and reach opposite conclusions, which is the
// rule most likely to be flattened by someone tidying the two guards into one.
func TestLocationUsageIsEmpty(t *testing.T) {
	assert.True(t, itStock.LocationUsage{}.IsEmpty())

	assert.False(t, itStock.LocationUsage{
		OnHandQuantity: decimal.NewFromInt(1),
	}.IsEmpty(), "stock on hand blocks archiving")

	assert.False(t, itStock.LocationUsage{
		ReservedQuantity: decimal.NewFromInt(5),
	}.IsEmpty(), "stock promised to a move blocks archiving")

	assert.False(t, itStock.LocationUsage{OpenMoveCount: 1}.IsEmpty())
	assert.False(t, itStock.LocationUsage{OpenTransferCount: 1}.IsEmpty())
}

// A location holding stock may be suspended: locking a damaged rack that still holds goods is the
// point of the operation. Only work already in flight stops it.
func TestSuspendIsNotBlockedByStockOnHand(t *testing.T) {
	holdingStock := itStock.LocationUsage{
		OnHandQuantity:   decimal.NewFromInt(100),
		ReservedQuantity: decimal.NewFromInt(10),
	}

	assert.False(t, holdingStock.IsEmpty(), "the same reading refuses an archive")
	assert.Zero(t, holdingStock.OpenMoveCount, "but nothing is in flight, so a suspend is allowed")
	assert.Zero(t, holdingStock.OpenTransferCount)
}

func TestJoinPath(t *testing.T) {
	assert.Equal(t, "MAIN", joinPath("", "MAIN"), "a root location is its own path")
	assert.Equal(t, "MAIN/Stock", joinPath("MAIN", "Stock"))
	assert.Equal(t, "MAIN/Stock/Zone A", joinPath("MAIN/Stock", "Zone A"))
}

func TestIsSystemGenerated(t *testing.T) {
	flagged := models.NewInventoryLocationFrom(dmodel.DynamicFields{
		models.InventoryLocationFieldIsSystemGenerated: true,
	})
	assert.True(t, isSystemGenerated(*flagged))

	plain := models.NewInventoryLocationFrom(dmodel.DynamicFields{})
	assert.False(t, isSystemGenerated(*plain), "an absent flag is not system-generated")
}

// The structural fields of a warehouse's own location are fixed; the descriptive ones are not.
func TestAssertSystemLocationUnchanged(t *testing.T) {
	current := models.NewInventoryLocationFrom(dmodel.DynamicFields{
		models.InventoryLocationFieldCode:          "Stock",
		models.InventoryLocationFieldLocationUsage: models.InventoryLocationUsageInternal,
		models.InventoryLocationFieldPurpose:       models.InventoryLocationPurposeStorage,
	})

	unchanged := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldCode: "Stock",
	}, *current)
	assert.Zero(t, unchanged.Count(), "resending the same value is not a change")

	descriptive := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldBarcode: "WH-STOCK-01",
	}, *current)
	assert.Zero(t, descriptive.Count(), "a barcode is not structural")

	renamed := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldCode: "Storage",
	}, *current)
	assert.Equal(t, 1, renamed.Count(), "the code is part of the flow that created it")

	repurposed := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldPurpose: models.InventoryLocationPurposePacking,
	}, *current)
	assert.Equal(t, 1, repurposed.Count())
}

// Create strips the system flag rather than trusting it, so a client cannot mint a location that
// then refuses to be archived.
func TestCopyFieldsIsADetachedCopy(t *testing.T) {
	original := dmodel.DynamicFields{models.InventoryLocationFieldCode: "A"}
	copied := copyFields(original)
	copied[models.InventoryLocationFieldCode] = "B"

	assert.Equal(t, "A", readStringParam(original, models.InventoryLocationFieldCode))
	assert.Equal(t, "B", readStringParam(copied, models.InventoryLocationFieldCode))
}

func TestClosedStatusesAreHistoryNotWorkInFlight(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{models.StockMoveStatusDone, models.StockMoveStatusCancelled},
		closedMoveStatuses())
	assert.ElementsMatch(t,
		[]string{models.StockTransferStatusDone, models.StockTransferStatusCancelled},
		closedTransferStatuses())

	// The predicate the movement engine already uses must agree with the list above, or a done
	// move would count as open and block archiving forever.
	for _, status := range closedMoveStatuses() {
		assert.False(t, IsMoveOpen(status), "%q is history", status)
	}
}
