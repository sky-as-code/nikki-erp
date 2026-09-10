package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Reservations held for a demand outside Inventory.
//
// The document is an ordinary stock transfer: it is created, confirmed and reserved through exactly
// the path a goods issue takes, and stops before validate. That reuse is the point — the quant row
// lock, the deterministic allocation order and the re-entrant outstanding-quantity arithmetic in
// stock_reservation.go are what make a hold correct under concurrency, and a second reservation
// mechanism beside them would be a second answer to "how much of this is spoken for".
//
// What is new is that the transfer remembers WHOSE demand it serves (source_type, source_id, and
// source_item_id per move), so the caller can come back and release or move the hold without having
// stored the transfer id, and so a partial outcome can be attributed line by line.

// ReserveForSource holds stock at one location for a caller's demand.
//
// Idempotent by the source pair: a caller retrying after a timeout is handed the hold it already
// created rather than a second one. That check and the creation are not atomic, which is deliberate
// and bounded — the loser of a race creates a duplicate transfer, and both are found and released by
// the same source pair. Reserving twice cannot over-claim stock, because each move only ever asks
// for its own outstanding quantity.
func (this *StockTransferDomainServiceImpl) ReserveForSource(
	ctx corectx.Context, request itStock.SourceReservationRequest,
) (*itStock.SourceReservationResult, error) {
	if result := assertSourceReservationWellFormed(request); result != nil {
		return result, nil
	}

	existing, err := this.findReservationBySource(ctx, request.SourceType, request.SourceId)
	if err != nil {
		return nil, err
	}
	if existing != "" {
		// A replay. Reserve again rather than answering from the row as it stands: the first call
		// may have died between creating the transfer and claiming the stock, and a hold that was
		// never taken must not be reported as held.
		return this.reserveExisting(ctx, existing)
	}

	params := dmodel.DynamicFields{
		models.StockTransferFieldOperationTypeId:   request.OperationTypeId,
		models.StockTransferFieldSourceLocationId:  request.LocationId,
		models.StockTransferFieldSourceType:        request.SourceType,
		models.StockTransferFieldSourceId:          request.SourceId,
		models.StockTransferFieldOrgId:             request.OrgId,
		models.StockTransferFieldReservationMethod: models.StockReservationMethodManual,
		models.StockTransferFieldOriginReference:   request.OriginReference,
	}
	if request.ExpiresAt != nil {
		params[models.StockTransferFieldExpiresAt] = *request.ExpiresAt
	}

	moves := make([]itStock.TransferMoveRequest, 0, len(request.Items))
	for _, item := range request.Items {
		moves = append(moves, itStock.TransferMoveRequest{
			ProductVariantId: item.ProductVariantId,
			UomId:            item.UomId,
			Quantity:         item.Quantity,
			SourceItemId:     item.SourceItemId,
		})
	}

	created, err := this.CreateWithMoves(ctx, params, moves)
	if err != nil {
		return nil, err
	}
	if created.ClientErrors.Count() > 0 || !created.HasData {
		return &itStock.SourceReservationResult{ClientErrors: created.ClientErrors}, nil
	}

	transferId := derefString(models.NewStockTransferFrom(created.Data).GetId())
	if transferId == "" {
		return nil, errors.New("the created reservation transfer carries no id")
	}
	return this.reserveExisting(ctx, transferId)
}

// reserveExisting confirms a draft hold and claims its stock, then reports what it actually got.
//
// A partial claim is a normal outcome of Reserve rather than an error, so this reads the shortage
// back and says so plainly: a caller needing all-or-nothing (a customer promised a specific kiosk)
// must be able to tell "held everything" from "held some", and a nil error says neither.
func (this *StockTransferDomainServiceImpl) reserveExisting(
	ctx corectx.Context, transferId string,
) (*itStock.SourceReservationResult, error) {
	confirmed, err := this.Confirm(ctx, transferId)
	if err != nil {
		return nil, err
	}
	// Confirm refuses a transfer that is already past draft, which a replay legitimately is; that
	// refusal is not a failure of the reservation, so it is not propagated.
	if confirmed.ClientErrors.Count() == 0 {
		if _, err := this.Reserve(ctx, transferId); err != nil {
			return nil, err
		}
	} else if _, err := this.Reserve(ctx, transferId); err != nil {
		return nil, err
	}

	availability, err := this.availabilityOf(ctx, transferId)
	if err != nil {
		return nil, err
	}
	return &itStock.SourceReservationResult{
		InventoryReference: transferId,
		FullyReserved:      availability != nil && !availability.HasShortage,
	}, nil
}

// availabilityOf re-reads what a transfer managed to claim.
func (this *StockTransferDomainServiceImpl) availabilityOf(
	ctx corectx.Context, transferId string,
) (*TransferAvailability, error) {
	result, err := this.CheckAvailability(ctx, transferId)
	if err != nil {
		return nil, err
	}
	if result == nil || !result.HasData {
		return nil, nil
	}
	report, ok := result.Data.(TransferAvailability)
	if !ok {
		return nil, errors.New("CheckAvailability returned an unexpected payload")
	}
	return &report, nil
}

// ReleaseReservationBySource gives back every hold taken for a demand.
//
// It releases by the source pair rather than by transfer id so a caller that has lost track of what
// it created — or that raced itself into two holds — still ends up with nothing reserved. Releasing
// a demand that holds nothing is success, not a refusal: the caller's intent is "this demand should
// hold no stock", and that is already true.
func (this *StockTransferDomainServiceImpl) ReleaseReservationBySource(
	ctx corectx.Context, sourceType string, sourceId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	transferIds, err := this.findAllReservationsBySource(ctx, sourceType, sourceId)
	if err != nil {
		return nil, err
	}
	for _, transferId := range transferIds {
		if _, err := this.Unreserve(ctx, transferId); err != nil {
			return nil, err
		}
	}
	return mutateOk(), nil
}

// findReservationBySource returns the first hold recorded for a demand, or an empty string.
func (this *StockTransferDomainServiceImpl) findReservationBySource(
	ctx corectx.Context, sourceType string, sourceId string,
) (string, error) {
	found, err := this.findAllReservationsBySource(ctx, sourceType, sourceId)
	if err != nil || len(found) == 0 {
		return "", err
	}
	return found[0], nil
}

// findAllReservationsBySource lists every open transfer raised for a demand. Done and cancelled
// documents are excluded: their stock has already moved or been given back, and unreserving one
// would be acting on history.
func (this *StockTransferDomainServiceImpl) findAllReservationsBySource(
	ctx corectx.Context, sourceType string, sourceId string,
) ([]string, error) {
	if sourceType == "" || sourceId == "" {
		return nil, nil
	}
	engine, err := repoFor(models.StockTransferSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.StockTransferFieldSourceType, dmodel.Equals, sourceType))
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.StockTransferFieldSourceId, dmodel.Equals, sourceId))

	found, err := engine.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  models.MaxTransferMoves,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findAllReservationsBySource")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	ids := make([]string, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		transfer := models.NewStockTransferFrom(item)
		if !IsTransferOpen(derefString(transfer.GetStatus())) {
			continue
		}
		ids = append(ids, derefString(transfer.GetId()))
	}
	return ids, nil
}

// ExpireLapsedReservations releases every hold whose expires_at has passed, and reports how many.
//
// It only ever releases stock: expiring a hold says the goods are sellable again, not that the
// demand behind it is cancelled or that anybody should be refunded — those are the owning module's
// decisions, and taking them here would refund a customer who is still standing at the machine.
//
// Each transfer is released in its own call rather than one sweeping transaction, so a single
// stubborn document cannot block the rest, and re-running after a partial failure is harmless.
func (this *StockTransferDomainServiceImpl) ExpireLapsedReservations(
	ctx corectx.Context, asOf time.Time, limit int,
) (int, error) {
	engine, err := repoFor(models.StockTransferSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.StockTransferFieldExpiresAt, dmodel.LessThan,
			model.ModelDateTime(asOf.UTC())))

	found, err := engine.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  limit,
	})
	if err != nil {
		return 0, errors.Wrap(err, "ExpireLapsedReservations")
	}
	if found == nil || !found.HasData {
		return 0, nil
	}

	released := 0
	for _, item := range found.Data.Items {
		transfer := models.NewStockTransferFrom(item)
		if !IsTransferOpen(derefString(transfer.GetStatus())) {
			continue
		}
		transferId := derefString(transfer.GetId())
		if _, err := this.Unreserve(ctx, transferId); err != nil {
			return released, err
		}
		if _, err := this.Cancel(ctx, transferId); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

// assertSourceReservationWellFormed refuses a request that could not name a hold, before anything is
// written. Without the source pair the transfer would be unfindable by the only key the caller has.
func assertSourceReservationWellFormed(
	request itStock.SourceReservationRequest,
) *itStock.SourceReservationResult {
	vErrs := ft.NewClientErrors()
	refuse := func(key, message string) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, key, message))
	}

	if request.SourceType == "" {
		refuse("stock_transfer.source_type_required",
			"a reservation held for another module must name what kind of demand it serves")
	}
	if request.SourceId == "" {
		refuse("stock_transfer.source_id_required",
			"a reservation held for another module must name the demand it serves")
	}
	if request.LocationId == "" {
		refuse("stock_transfer.source_location_required",
			"a reservation must name the location the stock is held at")
	}
	if request.OperationTypeId == "" {
		refuse("stock_transfer.operation_type_required",
			"a reservation must name the operation type it is raised under")
	}
	if len(request.Items) == 0 {
		refuse("stock_transfer.no_moves",
			"a reservation must ask for at least one item")
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return &itStock.SourceReservationResult{ClientErrors: *vErrs}
}
