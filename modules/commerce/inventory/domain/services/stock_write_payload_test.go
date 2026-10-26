package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// What the movement flows hand the repository. A repository write goes straight to the query builder,
// which refuses a string for a numeric column and reports it as client errors on a nil error, so a
// payload carrying text, or a caller dropping that refusal, makes the write silently not happen.

// moveWriteRepository stands in for the move repository. Its Update honours the etag the way the real
// one does: a stale etag matches no row and changes nothing, without an error.
type moveWriteRepository struct {
	composable.CrudRepository

	rows        []dmodel.DynamicFields
	inserts     []dmodel.DynamicFields
	refuseWrite bool
	etagSerial  int
}

func (this *moveWriteRepository) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	items := make([]dmodel.DynamicFields, 0, len(this.rows))
	for _, row := range this.rows {
		items = append(items, cloneFields(row))
	}
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: items, Total: len(items)},
		HasData: true,
	}, nil
}

func (this *moveWriteRepository) Insert(
	_ corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	if this.refuseWrite {
		return &dyn.OpResult[int]{ClientErrors: refusal()}, nil
	}
	this.inserts = append(this.inserts, data)
	return &dyn.OpResult[int]{Data: 1, HasData: true}, nil
}

func (this *moveWriteRepository) Update(
	_ corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if this.refuseWrite {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: refusal()}, nil
	}
	for _, row := range this.rows {
		if row[models.StockMoveFieldId] != data[models.StockMoveFieldId] ||
			row[basemodel.FieldEtag] != data[basemodel.FieldEtag] {
			continue
		}
		for key, value := range data {
			row[key] = value
		}
		this.etagSerial++
		row[basemodel.FieldEtag] = "etag-" + decimal.NewFromInt(int64(this.etagSerial)).String()
	}
	return &dyn.OpResult[dyn.MutateResultData]{HasData: true}, nil
}

func refusal() ft.ClientErrors {
	return ft.ClientErrors{*ft.NewValidationError(
		models.StockMoveFieldDemandQuantity, ft.ErrorKey("err_invalid_data_type"), "must be a numeric")}
}

func cloneFields(row dmodel.DynamicFields) dmodel.DynamicFields {
	clone := make(dmodel.DynamicFields, len(row))
	for key, value := range row {
		clone[key] = value
	}
	return clone
}

func sourceMove(status string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.StockMoveFieldId:                    "mv-1",
		models.StockMoveFieldTransferId:            "tr-1",
		models.StockMoveFieldProductVariantId:      testVariantAId,
		models.StockMoveFieldSourceLocationId:      testLocationAId,
		models.StockMoveFieldDestinationLocationId: testLocationBId,
		models.StockMoveFieldStatus:                status,
		models.StockMoveFieldOrgId:                 "org-1",
		basemodel.FieldEtag:                        "etag-0",
	}
}

func TestBackorderMovesCarryDecimalQuantities(t *testing.T) {
	moveRepo := &moveWriteRepository{}
	operation := &transferOperationContext{
		MoveRepo: moveRepo,
		Moves:    []dmodel.DynamicFields{sourceMove(models.StockMoveStatusDone)},
	}

	err := copyShortfallMoves(callerContext(), operation, []moveOutcome{outcome("100", "70")}, "tr-2")

	require.NoError(t, err)
	require.Len(t, moveRepo.inserts, 1)
	assert.Equal(t, dec(t, "30"), moveRepo.inserts[0][models.StockMoveFieldDemandQuantity])
	assert.Equal(t, dec(t, "30"), moveRepo.inserts[0][models.StockMoveFieldBaseDemandQuantity])
}

func TestBackorderMoveRefusedByTheRepositoryIsAnError(t *testing.T) {
	operation := &transferOperationContext{
		MoveRepo: &moveWriteRepository{refuseWrite: true},
		Moves:    []dmodel.DynamicFields{sourceMove(models.StockMoveStatusDone)},
	}

	err := copyShortfallMoves(callerContext(), operation, []moveOutcome{outcome("100", "70")}, "tr-2")

	assert.Error(t, err)
}

func TestReturnMovesCarryDecimalQuantities(t *testing.T) {
	moveRepo := &moveWriteRepository{}
	operation := &transferOperationContext{
		MoveRepo: moveRepo,
		Moves:    []dmodel.DynamicFields{sourceMove(models.StockMoveStatusDone)},
	}
	requested := []ReturnLineRequest{{MoveId: "mv-1", Quantity: dec(t, "4")}}

	err := insertReturnMoves(callerContext(), operation, "ret-1", requested)

	require.NoError(t, err)
	require.Len(t, moveRepo.inserts, 1)
	assert.Equal(t, dec(t, "4"), moveRepo.inserts[0][models.StockMoveFieldDemandQuantity])
	assert.Equal(t, dec(t, "4"), moveRepo.inserts[0][models.StockMoveFieldBaseDemandQuantity])
	assert.Equal(t, testLocationBId, moveRepo.inserts[0][models.StockMoveFieldSourceLocationId])
}

func TestReturnMoveRefusedByTheRepositoryIsAnError(t *testing.T) {
	operation := &transferOperationContext{
		MoveRepo: &moveWriteRepository{refuseWrite: true},
		Moves:    []dmodel.DynamicFields{sourceMove(models.StockMoveStatusDone)},
	}
	requested := []ReturnLineRequest{{MoveId: "mv-1", Quantity: dec(t, "4")}}

	assert.Error(t, insertReturnMoves(callerContext(), operation, "ret-1", requested))
}

func TestMoveStatusRefusedByTheRepositoryIsAnError(t *testing.T) {
	move := models.NewStockMoveFrom(sourceMove(models.StockMoveStatusAssigned))

	err := updateMoveStatus(callerContext(), &moveWriteRepository{refuseWrite: true}, *move,
		models.StockMoveStatusConfirmed)

	assert.Error(t, err)
}

func TestCancellingAReservedTransferCancelsItsMoves(t *testing.T) {
	// Releasing the reservation rewrites the move (and its etag) before the cancel; the cancel must
	// write against that state, not the one loaded with the transfer.
	moveRepo := &moveWriteRepository{rows: []dmodel.DynamicFields{sourceMove(models.StockMoveStatusAssigned)}}
	operation := &transferOperationContext{
		MoveRepo:     moveRepo,
		MoveLineRepo: &stubRowRepository{},
		Moves:        []dmodel.DynamicFields{sourceMove(models.StockMoveStatusAssigned)},
	}

	err := releaseThenCancelMoves(callerContext(), operation, "tr-1")

	require.NoError(t, err)
	assert.Equal(t, models.StockMoveStatusCancelled, moveRepo.rows[0][models.StockMoveFieldStatus])
}
