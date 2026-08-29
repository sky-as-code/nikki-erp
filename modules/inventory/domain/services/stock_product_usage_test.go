package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// What Stock reports when Product asks whether a variant can be retired. The distinction is history
// versus work in flight: a variant that has BEEN moved archives fine, one that IS being moved does
// not.

const testTransferId = "01TRANSFER00000000000000AA"

func moveRow(variantId, status, transferId string) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		models.StockMoveFieldProductVariantId: variantId,
		models.StockMoveFieldStatus:           status,
	}
	if transferId != "" {
		row[models.StockMoveFieldTransferId] = transferId
	}
	return row
}

func readUsage(t *testing.T, variantId string) itStock.ProductUsage {
	t.Helper()

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetProductUsage(
		callerContext(), itStock.GetProductUsageQuery{VariantId: variantId})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result.Data.Usage
}

// Stock on the shelf blocks an archive, and IsEmpty is what says so.
func TestProductUsageReportsOnHandStock(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {quantRow(testVariantAId, testLocationAId, 100, 0)},
	})

	usage := readUsage(t, testVariantAId)

	assert.Equal(t, "100", usage.OnHandQuantity.String())
	assert.False(t, usage.IsEmpty(), "a variant with stock cannot be archived")
}

// A variant referenced only by completed movement archives fine: history keeps resolving it, so
// counting it would block a safe archive forever.
func TestProductUsageIgnoresCompletedMovement(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {quantRow(testVariantAId, testLocationAId, 0, 0)},
		models.StockMoveSchemaName: {
			moveRow(testVariantAId, models.StockMoveStatusDone, testTransferId),
			moveRow(testVariantAId, models.StockMoveStatusCancelled, testTransferId),
		},
	})

	usage := readUsage(t, testVariantAId)

	assert.Equal(t, 0, usage.OpenMoveCount, "done and cancelled moves are history")
	assert.Equal(t, 0, usage.OpenTransferCount)
	assert.True(t, usage.IsEmpty(), "historical-only stock does not block archiving")
}

// Work in flight does block it, and a transfer carrying several lines of one product counts once.
func TestProductUsageCountsOpenWorkAndDedupesTransfers(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockMoveSchemaName: {
			moveRow(testVariantAId, models.StockMoveStatusConfirmed, testTransferId),
			moveRow(testVariantAId, models.StockMoveStatusAssigned, testTransferId),
		},
	})

	usage := readUsage(t, testVariantAId)

	assert.Equal(t, 2, usage.OpenMoveCount, "each open move counts")
	assert.Equal(t, 1, usage.OpenTransferCount, "two lines of one transfer are one transfer")
	assert.False(t, usage.IsEmpty())
}

// Reservation alone blocks an archive even when nothing is physically left to take away.
func TestProductUsageReportsReservationSeparately(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {quantRow(testVariantAId, testLocationAId, 5, 5)},
	})

	usage := readUsage(t, testVariantAId)

	assert.Equal(t, "5", usage.ReservedQuantity.String())
	assert.False(t, usage.IsEmpty())
}

// The batch read is what the template guard uses, so every requested variant must come back —
// including the ones holding nothing.
func TestProductUsageBatchReportsEveryRequestedVariant(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {quantRow(testVariantBId, testLocationAId, 10, 0)},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetProductUsageBatch(callerContext(), itStock.GetProductUsageBatchQuery{
		VariantIds: []string{testVariantAId, testVariantBId},
	})

	require.NoError(t, err)
	require.Len(t, result.Data.Usages, 2)
	assert.True(t, result.Data.Usages[testVariantAId].IsEmpty(), "a clean variant is reported clean")
	assert.False(t, result.Data.Usages[testVariantBId].IsEmpty(), "the one holding stock is not")
}

// A variant id is required: an empty one would otherwise summarise across the whole catalogue and
// report a blocked archive for a product that is perfectly clean.
func TestProductUsageRejectsAnEmptyVariantId(t *testing.T) {
	service := &StockQuantDomainServiceImpl{}

	_, err := service.GetProductUsage(callerContext(), itStock.GetProductUsageQuery{VariantId: ""})

	assert.Error(t, err)
}
