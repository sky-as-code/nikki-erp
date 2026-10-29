package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
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

func TestBuildLocationCanonicalName(t *testing.T) {
	assert.Equal(t, "MAIN", BuildLocationCanonicalName("", "MAIN"), "a location without a parent is its own name")
	assert.Equal(t, "MAIN / Zone A", BuildLocationCanonicalName("MAIN", "Zone A"))
	assert.Equal(t, "MAIN / Zone A / Rack 1", BuildLocationCanonicalName("MAIN / Zone A", "Rack 1"))
}

// A location inherits its parent's warehouse; without a parent it has none, unless it is a
// warehouse's root, which belongs to the warehouse managing it.
func TestDeriveLocationNode(t *testing.T) {
	parent := models.NewInventoryLocationFrom(dmodel.DynamicFields{
		models.InventoryLocationFieldName:           "MAIN",
		models.InventoryLocationFieldCanonicalName:  "MAIN",
		models.InventoryLocationFieldWarehouseId:    "01WAREHOUSE",
		models.InventoryLocationFieldHierarchyDepth: int32(0),
	})
	child := models.NewInventoryLocationFrom(dmodel.DynamicFields{})
	node := deriveLocationNode(*child, parent, "Zone A")
	assert.Equal(t, "MAIN / Zone A", node.canonical)
	assert.Equal(t, "01WAREHOUSE", node.warehouseId)
	assert.Equal(t, 1, node.depth)

	orphan := deriveLocationNode(*child, nil, "Zone A")
	assert.Equal(t, "Zone A", orphan.canonical)
	assert.Empty(t, orphan.warehouseId, "clearing the parent clears the warehouse")
	assert.Zero(t, orphan.depth)

	root := models.NewInventoryLocationFrom(dmodel.DynamicFields{
		"managed_by_type": models.InventoryLocationManagedByWarehouse,
		"managed_by_id":   "01WAREHOUSE",
	})
	assert.Equal(t, "01WAREHOUSE", deriveLocationNode(*root, nil, "MAIN").warehouseId)
}

func TestApplyLocationNodeWritesNullWarehouse(t *testing.T) {
	fields := dmodel.DynamicFields{}
	applyLocationNode(fields, locationNode{canonical: "A"})
	value, present := fields[models.InventoryLocationFieldWarehouseId]
	assert.True(t, present, "a cleared warehouse must be written, not skipped")
	assert.Nil(t, value)
}

func TestRejectDerivedLocationFields(t *testing.T) {
	vErrs := ft.NewClientErrors()
	rejectDerivedLocationFields(dmodel.DynamicFields{models.InventoryLocationFieldCanonicalName: "x"}, vErrs)
	assert.Equal(t, 1, vErrs.Count())
}

// The type of a warehouse's root is fixed; its name is not.
func TestAssertWarehouseRootUnchanged(t *testing.T) {
	root := models.NewInventoryLocationFrom(dmodel.DynamicFields{
		models.InventoryLocationFieldLocationType: models.InventoryLocationTypeGroup,
	})
	assert.Zero(t, assertWarehouseRootUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldName:         "Renamed",
		models.InventoryLocationFieldLocationType: models.InventoryLocationTypeGroup,
	}, *root).Count())
	assert.Equal(t, 1, assertWarehouseRootUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldLocationType: models.InventoryLocationTypeInternal,
	}, *root).Count())
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
		models.InventoryLocationFieldParentLocationId: "01ROOT",
		models.InventoryLocationFieldLocationType:     models.InventoryLocationTypeInternal,
	})

	unchanged := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldParentLocationId: "01ROOT",
	}, *current)
	assert.Zero(t, unchanged.Count(), "resending the same value is not a change")

	descriptive := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldBarcode: "WH-STOCK-01",
	}, *current)
	assert.Zero(t, descriptive.Count(), "a barcode is not structural")

	renamed := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldName: "Storage",
	}, *current)
	assert.Zero(t, renamed.Count(), "a system location may be renamed")

	moved := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldParentLocationId: "01ELSEWHERE",
	}, *current)
	assert.Equal(t, 1, moved.Count(), "the parent is part of the flow that created it")

	retyped := assertSystemLocationUnchanged(dmodel.DynamicFields{
		models.InventoryLocationFieldLocationType: models.InventoryLocationTypeGroup,
	}, *current)
	assert.Equal(t, 1, retyped.Count())
}

// Create strips the system flag rather than trusting it, so a client cannot mint a location that
// then refuses to be archived.
func TestCopyFieldsIsADetachedCopy(t *testing.T) {
	original := dmodel.DynamicFields{models.InventoryLocationFieldName: "A"}
	copied := copyFields(original)
	copied[models.InventoryLocationFieldName] = "B"

	assert.Equal(t, "A", readStringParam(original, models.InventoryLocationFieldName))
	assert.Equal(t, "B", readStringParam(copied, models.InventoryLocationFieldName))
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
