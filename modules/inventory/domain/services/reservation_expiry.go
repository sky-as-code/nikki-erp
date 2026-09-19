package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Expiry bookkeeping (CR-INV-SALES-WH-RESERVATION §5.3).
//
// A reservation lapses the instant the clock passes reserved_until, and every availability figure
// already says so through IsReservationEffective. What is written here is the record of that
// fact: the row is closed as released with reason "expired", released_at is the deadline itself,
// and expiry_recorded_at is when the bookkeeping caught up. Nothing about availability depends on
// this having run, which is why there is no timer per reservation and the sweep may be hours late
// without a single unit being over-committed.

// MaterializeExpiredInScope closes the lapsed rows of the given scopes. Called by every operation
// that already holds the scopes' guards, so the expiry event is written by whoever touches the
// scope first rather than waiting for the sweep.
func MaterializeExpiredInScope(ctx corectx.Context, keys []GuardKey, now time.Time) (int, error) {
	count := 0
	for _, key := range sortGuardKeys(keys) {
		reservations, err := activeReservationsOfScope(ctx, key)
		if err != nil {
			return count, err
		}
		for _, reservation := range reservations {
			if IsReservationEffective(reservation, now) {
				continue
			}
			if err := materializeExpiry(ctx, reservation, now); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// materializeExpiry closes one lapsed row and announces it, with the deadline as the effective
// time and now as when it was recorded.
func materializeExpiry(ctx corectx.Context, reservation models.StockReservation, now time.Time) error {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return err
	}
	deadline := reservation.GetReservedUntil()
	if deadline == nil {
		return errors.New("materializeExpiry called on a reservation without a deadline")
	}
	remaining := reservation.RemainingQuantity()
	released := derefDecimal(reservation.GetReleasedQuantity()).Add(remaining)

	if _, err := engine.Update(ctx, dmodel.DynamicFields{
		models.StockReservationFieldId:               derefId(reservation.GetId()),
		models.StockReservationFieldReleasedQuantity: released,
		models.StockReservationFieldStatus:           models.StockReservationStatusReleased,
		models.StockReservationFieldReleaseReason:    models.StockReservationReleaseReasonExpired,
		models.StockReservationFieldReleasedAt:       *deadline,
		models.StockReservationFieldExpiryRecordedAt: model.WrapModelDateTime(now),
	}); err != nil {
		return errors.Wrap(err, "materializeExpiry")
	}

	_, err = RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventInventoryReservationExpired,
		AggregateId: derefId(reservation.GetId()),
		OrgId:       derefId(reservation.GetOrgId()),
		Payload: map[string]any{
			"reservation_id":     derefId(reservation.GetId()),
			"released_quantity":  remaining.String(),
			"effective_at":       deadline.GoTime().UTC().Format(time.RFC3339),
			"recorded_at":        now.UTC().Format(time.RFC3339),
			"warehouse_id":       derefId(reservation.GetWarehouseId()),
			"product_variant_id": derefId(reservation.GetProductVariantId()),
		},
		OccurredAt: deadline.GoTime().Unix(),
	})
	return err
}

// ExpireLapsedWarehouseReservations is the sweep's entry: it finds up to limit stored-active rows
// whose deadline has passed and materializes each under its scope's guard, one transaction per
// row so a stubborn row cannot block the rest.
func (this *StockReservationDomainServiceImpl) ExpireLapsedWarehouseReservations(
	ctx corectx.Context, asOf time.Time, limit int,
) (int, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return 0, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldStatus, dmodel.Equals, models.StockReservationStatusActive),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldReservedUntil, dmodel.LessEqual, model.ModelDateTime(asOf.UTC())),
	)
	found, err := engine.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: limit})
	if err != nil {
		return 0, errors.Wrap(err, "ExpireLapsedWarehouseReservations")
	}
	if found == nil || !found.HasData {
		return 0, nil
	}

	expired := 0
	for _, row := range found.Data.Items {
		key := guardKeyOf(*models.NewStockReservationFrom(row))
		err := withReservationTransaction(ctx, func(tranxCtx corectx.Context) error {
			guardRepo, err := repoFor(models.WarehouseProductGuardSchemaName)
			if err != nil {
				return err
			}
			if _, err := LockGuardsForUpdate(tranxCtx, guardRepo.GetBaseRepo(), []GuardKey{key}); err != nil {
				return err
			}
			now, err := DbNowUnderLock(tranxCtx, guardRepo.GetBaseRepo())
			if err != nil {
				return err
			}
			count, err := MaterializeExpiredInScope(tranxCtx, []GuardKey{key}, now)
			expired += count
			return err
		})
		if err != nil {
			return expired, err
		}
	}
	if expired > 0 {
		DrainOutboxNow(ctx)
	}
	return expired, nil
}

// ResolveWarehouseOfLocation answers the warehouse a location belongs to, for a caller that knows
// only its point of sale's location.
func (this *StockReservationDomainServiceImpl) ResolveWarehouseOfLocation(
	ctx corectx.Context, orgId, locationId model.Id,
) (model.Id, error) {
	location, err := findRecord(ctx, models.InventoryLocationSchemaName, models.InventoryLocationFieldId, string(locationId))
	if err != nil {
		return "", err
	}
	if location == nil || stringOf(location, models.InventoryLocationFieldOrgId) != string(orgId) {
		return "", nil
	}
	return model.Id(stringOf(location, models.InventoryLocationFieldWarehouseId)), nil
}

// ReservationsOfSource lists a demand revision's rows as the caller sees them now.
func (this *StockReservationDomainServiceImpl) ReservationsOfSource(
	ctx corectx.Context, orgId model.Id, sourceModule, sourceType, sourceId string, revision int32,
) ([]itStock.ReservedLine, error) {
	if revision <= 0 {
		revision = 1
	}
	reservations, err := reservationsOfSource(ctx, orgId, sourceModule, sourceType, sourceId, revision)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	lines := make([]itStock.ReservedLine, 0, len(reservations))
	for _, reservation := range reservations {
		lines = append(lines, reservedLineOf(reservation, now))
	}
	return lines, nil
}
