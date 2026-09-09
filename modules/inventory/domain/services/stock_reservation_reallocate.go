package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// ReallocateReservation moves a demand's hold from wherever it is to a new location.
//
// The ordering is the whole point and is not negotiable: claim at the destination FIRST, and only
// once that has fully succeeded give back the origin. Doing it the other way — release, then
// reserve — opens a window in which the customer holds nothing, and a concurrent sale can take the
// goods they were already promised. If the destination cannot supply everything, the transaction
// rolls back and the customer keeps exactly the hold they had.
//
// Everything happens in one transaction for the same reason. Two transactions would commit the new
// claim before the old release, and a crash between them would leave the demand holding stock in two
// places, quietly making both unsellable.
func (this *StockTransferDomainServiceImpl) ReallocateReservation(
	ctx corectx.Context, request itStock.ReservationReallocationRequest,
) (*itStock.SourceReservationResult, error) {
	if result := assertReallocationWellFormed(request); result != nil {
		return result, nil
	}

	var outcome *itStock.SourceReservationResult

	err := withTransferTransaction(ctx, func(tranxCtx corectx.Context) error {
		// Both sides are locked before either is written, in one deterministic global order, so two
		// reallocations swapping the same pair of locations queue behind each other instead of each
		// holding half of what the other needs. Ordering by location alone would be enough here, but
		// the sort is over the whole key so a future caller moving several variants at once inherits
		// the same discipline.
		if err := lockBothSides(tranxCtx, request); err != nil {
			return err
		}

		previous, err := this.findAllReservationsBySource(tranxCtx, request.SourceType, request.SourceId)
		if err != nil {
			return err
		}

		// The new hold is a fresh document rather than an edit of the old one: a transfer names its
		// source location on the header and on every move, so "moving" one would rewrite the record
		// of where the goods were promised from. Two documents leave the history readable — held
		// here, then held there.
		claimed, err := this.ReserveForSource(tranxCtx, itStock.SourceReservationRequest{
			SourceType:      request.SourceType,
			SourceId:        request.SourceId,
			OrgId:           request.OrgId,
			LocationId:      request.ToLocationId,
			OperationTypeId: request.OperationTypeId,
			OriginReference: request.OriginReference,
			ExpiresAt:       request.ExpiresAt,
			Items:           request.Items,
		})
		if err != nil {
			return err
		}
		if claimed.ClientErrors.Count() > 0 {
			outcome = claimed
			return errRollbackReallocation
		}
		if !claimed.FullyReserved {
			outcome = &itStock.SourceReservationResult{
				ClientErrors: reallocationShortage(request.ToLocationId),
			}
			return errRollbackReallocation
		}

		// Only now, with the destination fully claimed inside this same transaction, is the old hold
		// given back. A failure here rolls back the new claim too, which is correct: the demand ends
		// up exactly where it started.
		for _, transferId := range previous {
			if transferId == claimed.InventoryReference {
				continue
			}
			if _, err := this.Unreserve(tranxCtx, transferId); err != nil {
				return err
			}
			if _, err := this.Cancel(tranxCtx, transferId); err != nil {
				return err
			}
		}

		outcome = claimed
		return nil
	})

	// The sentinel is how a refusal escapes the transaction with the writes undone. It is not a
	// fault: outcome already carries the reason, and a Go error here would answer 500 for something
	// the caller can fix by choosing another location.
	if errors.Is(err, errRollbackReallocation) {
		return outcome, nil
	}
	if err != nil {
		return nil, err
	}
	return outcome, nil
}

// errRollbackReallocation unwinds the transaction while keeping the refusal that caused it.
var errRollbackReallocation = errors.New("reallocation refused; rolling back")

// lockBothSides takes the row locks for every variant at both the origin and the destination before
// anything is written, in a single deterministic order.
//
// The origin is not named by the caller — it is wherever the demand's current moves point — so it is
// read from those moves rather than trusted from the request, where it could disagree with reality.
func lockBothSides(
	ctx corectx.Context, request itStock.ReservationReallocationRequest,
) error {
	quantEngine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return err
	}

	keys := make([]QuantLockKey, 0, len(request.Items)*2)
	for _, item := range request.Items {
		keys = append(keys, QuantLockKey{
			OrgId:            request.OrgId,
			ProductVariantId: item.ProductVariantId,
			LocationId:       request.ToLocationId,
		})
	}

	origins, err := originLocationsOf(ctx, request)
	if err != nil {
		return err
	}
	for _, item := range request.Items {
		for _, location := range origins {
			keys = append(keys, QuantLockKey{
				OrgId:            request.OrgId,
				ProductVariantId: item.ProductVariantId,
				LocationId:       location,
			})
		}
	}

	sortQuantLockKeys(keys)

	seen := make(map[QuantLockKey]bool, len(keys))
	for _, key := range keys {
		if seen[key] {
			continue
		}
		seen[key] = true
		if _, err := LockQuantsForUpdate(
			ctx, quantEngine.ResourceRepository().GetBaseRepo(), key); err != nil {
			return err
		}
	}
	return nil
}

// sortQuantLockKeys puts a set of lock keys into one total order, so that any two callers touching
// an overlapping pair of rows take them in the same sequence and queue rather than deadlock. Two
// reallocations moving stock between the same locations in opposite directions are exactly that
// case. Location first, then variant: the order must not depend on which way the goods are going.
func sortQuantLockKeys(keys []QuantLockKey) {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].LocationId != keys[j].LocationId {
			return keys[i].LocationId < keys[j].LocationId
		}
		return keys[i].ProductVariantId < keys[j].ProductVariantId
	})
}

// originLocationsOf reads where a demand's stock is currently held, from the transfers that hold it.
func originLocationsOf(
	ctx corectx.Context, request itStock.ReservationReallocationRequest,
) ([]string, error) {
	engine, err := engineFor(models.StockTransferSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.StockTransferFieldSourceType, dmodel.Equals, request.SourceType))
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.StockTransferFieldSourceId, dmodel.Equals, request.SourceId))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  models.MaxTransferMoves,
	})
	if err != nil {
		return nil, errors.Wrap(err, "originLocationsOf")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	seen := map[string]bool{}
	locations := make([]string, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		transfer := models.NewStockTransferFrom(item)
		if !IsTransferOpen(derefString(transfer.GetStatus())) {
			continue
		}
		location := derefString(transfer.GetSourceLocationId())
		if location == "" || seen[location] {
			continue
		}
		seen[location] = true
		locations = append(locations, location)
	}
	return locations, nil
}

// reallocationShortage is the refusal when the destination cannot hold everything. It names the
// location so a client can offer the customer a different one rather than merely reporting failure.
func reallocationShortage(locationId string) ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockTransferSchemaName,
		"stock_transfer.reallocation_insufficient_stock",
		"location '"+locationId+"' cannot hold all of the requested quantity; the existing "+
			"reservation is unchanged",
	))
	return *vErrs
}

// assertReallocationWellFormed refuses a request that names no demand or no destination, before any
// lock is taken.
func assertReallocationWellFormed(
	request itStock.ReservationReallocationRequest,
) *itStock.SourceReservationResult {
	vErrs := ft.NewClientErrors()
	refuse := func(key, message string) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, key, message))
	}

	if request.SourceType == "" || request.SourceId == "" {
		refuse("stock_transfer.source_required",
			"a reallocation must name the demand whose reservation is moving")
	}
	if request.ToLocationId == "" {
		refuse("stock_transfer.destination_location_required",
			"a reallocation must name the location the stock should be held at instead")
	}
	if request.OperationTypeId == "" {
		refuse("stock_transfer.operation_type_required",
			"a reallocation must name the operation type the new hold is raised under")
	}
	if len(request.Items) == 0 {
		refuse("stock_transfer.no_moves",
			"a reallocation must name at least one item to move")
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return &itStock.SourceReservationResult{ClientErrors: *vErrs}
}
