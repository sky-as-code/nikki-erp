package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The Stock side of the location lifecycle contract, on the quant service because that is where the
// balances live. Warehouse Management calls it before suspending or archiving a location; it reads
// and reports, never changing anything.

var _ itStock.LocationUsageReadService = (*StockQuantDomainServiceImpl)(nil)

// GetLocationUsage reports what Stock holds at one location: quantities off the quants, and counts
// off moves and transfers using the repository's Total rather than fetching rows.
func (this *StockQuantDomainServiceImpl) GetLocationUsage(
	ctx corectx.Context, query itStock.GetLocationUsageQuery,
) (*itStock.GetLocationUsageResult, error) {
	if query.LocationId == "" {
		return nil, errors.New("GetLocationUsage requires a location id")
	}

	onHand, reserved, err := sumQuantitiesAtLocation(ctx, query.LocationId)
	if err != nil {
		return nil, err
	}

	openMoves, err := countOpenMovesAtLocation(ctx, query.LocationId)
	if err != nil {
		return nil, err
	}

	openTransfers, err := countOpenTransfersAtLocation(ctx, query.LocationId)
	if err != nil {
		return nil, err
	}

	return &itStock.GetLocationUsageResult{
		HasData: true,
		Data: itStock.GetLocationUsageResultData{
			Usage: itStock.LocationUsage{
				OnHandQuantity:    onHand,
				ReservedQuantity:  reserved,
				OpenMoveCount:     openMoves,
				OpenTransferCount: openTransfers,
			},
		},
	}, nil
}

// sumQuantitiesAtLocation totals the on-hand and reserved quantities of every quant at a location.
// Summed in memory because the dynamic-model layer offers no aggregation, and read through to the
// last page rather than truncated: a partial total would report an empty location that is not.
func sumQuantitiesAtLocation(
	ctx corectx.Context, locationId string,
) (decimal.Decimal, decimal.Decimal, error) {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldLocationId, dmodel.Equals, locationId),
	)

	onHand, reserved := decimal.Zero, decimal.Zero
	for page := 0; ; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  usageScanPageSize,
		})
		if err != nil {
			return decimal.Zero, decimal.Zero, errors.Wrap(err, "sumQuantitiesAtLocation")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			quant := models.NewStockQuantFrom(row)
			onHand = onHand.Add(derefDecimal(quant.GetOnHandQuantity()))
			reserved = reserved.Add(derefDecimal(quant.GetReservedQuantity()))
		}

		if len(found.Data.Items) < usageScanPageSize {
			break
		}
	}
	return onHand, reserved, nil
}

// countOpenMovesAtLocation counts the moves still in flight through a location, in either
// direction. Done and cancelled moves are excluded: an archived location still resolves for the
// records that name it, so history never blocks a lifecycle change.
func countOpenMovesAtLocation(ctx corectx.Context, locationId string) (int, error) {
	engine, err := engineFor(models.StockMoveSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(
				models.StockMoveFieldSourceLocationId, dmodel.Equals, locationId),
			*dmodel.NewSearchNode().NewCondition(
				models.StockMoveFieldDestinationLocationId, dmodel.Equals, locationId),
		),
		*dmodel.NewSearchNode().NewCondition(
			models.StockMoveFieldStatus, dmodel.NotIn, closedMoveStatuses()),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countOpenMovesAtLocation")
}

// countOpenTransfersAtLocation counts the transfers still in flight through a location.
func countOpenTransfersAtLocation(ctx corectx.Context, locationId string) (int, error) {
	engine, err := engineFor(models.StockTransferSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(
				models.StockTransferFieldSourceLocationId, dmodel.Equals, locationId),
			*dmodel.NewSearchNode().NewCondition(
				models.StockTransferFieldDestinationLocationId, dmodel.Equals, locationId),
		),
		*dmodel.NewSearchNode().NewCondition(
			models.StockTransferFieldStatus, dmodel.NotIn, closedTransferStatuses()),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countOpenTransfersAtLocation")
}
