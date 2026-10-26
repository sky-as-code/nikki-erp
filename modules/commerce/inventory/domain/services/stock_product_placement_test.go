package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Where a variant's stock sits, grouped two ways: a suspended place still shows its contents, and
// stock outside any warehouse is not attributed to one.

const testExternalLocationId = "01LOCATIONEXT0000000000000"

func locationRow(id, status, warehouseId, code string) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		models.InventoryLocationFieldId:            id,
		models.InventoryLocationFieldLocationUsage: models.InventoryLocationUsageInternal,
		models.InventoryLocationFieldStatus:        status,
		models.InventoryLocationFieldCode:          code,
	}
	if warehouseId != "" {
		row[models.InventoryLocationFieldWarehouseId] = warehouseId
	}
	return row
}

func warehouseRow(id, status, code string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.WarehouseFieldId:     id,
		models.WarehouseFieldCode:   code,
		models.WarehouseFieldStatus: status,
	}
}

// A suspended location keeps its stock and keeps being listed: suspension governs what may be
// chosen for new work, not whether what is already there exists.
func TestStockByLocationShowsSuspendedLocations(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {quantRow(testVariantAId, testLocationAId, 10, 0)},
		models.InventoryLocationSchemaName: {
			locationRow(testLocationAId, models.InventoryLocationStatusSuspended, testWarehouseId, "A01"),
		},
		models.WarehouseSchemaName: {warehouseRow(testWarehouseId, models.WarehouseStatusActive, "MAIN")},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetStockByLocation(
		callerContext(), itStock.GetStockByLocationQuery{VariantId: testVariantAId})

	require.NoError(t, err)
	require.Len(t, result.Data.Rows, 1)
	assert.Equal(t, "10", result.Data.Rows[0].OnHand.String())
	assert.Equal(t, models.InventoryLocationStatusSuspended, result.Data.Rows[0].LocationStatus,
		"the status travels with the row so the UI can badge it rather than hide it")
}

// Stock at a vendor, customer, transit or loss location belongs to no warehouse. It must be its
// own group rather than being folded into whichever warehouse sorts first.
func TestStockByWarehouseKeepsWarehouselessStockSeparate(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 70, 0),
			quantRow(testVariantAId, testExternalLocationId, 30, 0),
		},
		models.InventoryLocationSchemaName: {
			locationRow(testLocationAId, models.InventoryLocationStatusActive, testWarehouseId, "A01"),
			locationRow(testExternalLocationId, models.InventoryLocationStatusActive, "", "TRANSIT"),
		},
		models.WarehouseSchemaName: {warehouseRow(testWarehouseId, models.WarehouseStatusActive, "MAIN")},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetStockByWarehouse(
		callerContext(), itStock.GetStockByWarehouseQuery{VariantId: testVariantAId})

	require.NoError(t, err)
	require.Len(t, result.Data.Rows, 2)

	assert.Equal(t, "70", result.Data.Rows[0].OnHand.String(), "largest holding sorts first")
	require.NotNil(t, result.Data.Rows[0].WarehouseId)
	assert.Equal(t, "MAIN", result.Data.Rows[0].WarehouseCode)

	assert.Nil(t, result.Data.Rows[1].WarehouseId,
		"stock outside a warehouse must not be attributed to one")
	assert.Equal(t, "30", result.Data.Rows[1].OnHand.String())
}

// The warehouse rollup adds up the locations inside it.
func TestStockByWarehouseSumsTheLocationsWithinIt(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 30, 5),
			quantRow(testVariantAId, testLocationBId, 40, 0),
		},
		models.InventoryLocationSchemaName: {
			locationRow(testLocationAId, models.InventoryLocationStatusActive, testWarehouseId, "A01"),
			locationRow(testLocationBId, models.InventoryLocationStatusActive, testWarehouseId, "A02"),
		},
		models.WarehouseSchemaName: {warehouseRow(testWarehouseId, models.WarehouseStatusActive, "MAIN")},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetStockByWarehouse(
		callerContext(), itStock.GetStockByWarehouseQuery{VariantId: testVariantAId})

	require.NoError(t, err)
	require.Len(t, result.Data.Rows, 1, "two locations in one warehouse make one row")
	assert.Equal(t, "70", result.Data.Rows[0].OnHand.String())
	assert.Equal(t, "65", result.Data.Rows[0].Available.String())
}

// Narrowing to a warehouse is the drill-down from the rollup, so it must exclude everything else.
func TestStockByLocationNarrowsToOneWarehouse(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 10, 0),
			quantRow(testVariantAId, testExternalLocationId, 99, 0),
		},
		models.InventoryLocationSchemaName: {
			locationRow(testLocationAId, models.InventoryLocationStatusActive, testWarehouseId, "A01"),
			locationRow(testExternalLocationId, models.InventoryLocationStatusActive, "", "TRANSIT"),
		},
		models.WarehouseSchemaName: {warehouseRow(testWarehouseId, models.WarehouseStatusActive, "MAIN")},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetStockByLocation(callerContext(), itStock.GetStockByLocationQuery{
		VariantId:   testVariantAId,
		WarehouseId: testWarehouseId,
	})

	require.NoError(t, err)
	require.Len(t, result.Data.Rows, 1)
	assert.Equal(t, "A01", result.Data.Rows[0].LocationCode)
}

// A place holding nothing is not somewhere the product is. Listing it would pad the page with
// rows a reader has to skip past.
func TestStockByLocationOmitsEmptiedLocations(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 12, 0),
			quantRow(testVariantAId, testLocationBId, 0, 0),
		},
		models.InventoryLocationSchemaName: {
			locationRow(testLocationAId, models.InventoryLocationStatusActive, testWarehouseId, "A01"),
			locationRow(testLocationBId, models.InventoryLocationStatusActive, testWarehouseId, "A02"),
		},
		models.WarehouseSchemaName: {warehouseRow(testWarehouseId, models.WarehouseStatusActive, "MAIN")},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetStockByLocation(
		callerContext(), itStock.GetStockByLocationQuery{VariantId: testVariantAId})

	require.NoError(t, err)
	require.Len(t, result.Data.Rows, 1)
	assert.Equal(t, "A01", result.Data.Rows[0].LocationCode)
}

func TestPlacementReadsRejectAnEmptyVariantId(t *testing.T) {
	service := &StockQuantDomainServiceImpl{}

	_, byWarehouseErr := service.GetStockByWarehouse(
		callerContext(), itStock.GetStockByWarehouseQuery{VariantId: ""})
	_, byLocationErr := service.GetStockByLocation(
		callerContext(), itStock.GetStockByLocationQuery{VariantId: ""})

	assert.Error(t, byWarehouseErr)
	assert.Error(t, byLocationErr)
}
