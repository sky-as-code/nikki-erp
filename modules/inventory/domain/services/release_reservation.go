package services

import (
	"maps"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Releasing a reservation (CR-INV-SALES-WH-RESERVATION §5.2).
//
// A release gives back what has not been consumed and moves nothing: the goods never left. It is
// idempotent by construction — a reservation already released, lapsed or fully consumed has
// nothing left to give back, so a second call succeeds with released_now = 0 and writes no second
// event. A partial release is the exception: it must carry a key, because "release 2 more" said
// twice is not the same as said once.
//
// Whether the quantity may still be dispensed by an execution in flight is the owning module's
// knowledge, not Inventory's: the selling module refuses to release while a kiosk may still act.

const ReasonPartialReleaseNeedsKey = "stock_reservation.partial_release_needs_key"

// ReleaseReservation gives back the remainder, or the requested part of it.
func (this *StockReservationDomainServiceImpl) ReleaseReservation(
	ctx corectx.Context, request itStock.ReleaseReservationRequest,
) (*itStock.ReleaseReservationResult, error) {
	vErrs := ft.NewClientErrors()
	if request.OrgId == "" || request.ReservationId == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
			"org_id and reservation id are required"))
	}
	if request.Quantity.IsNegative() {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
			"quantity must not be negative"))
	}
	if request.Quantity.IsPositive() && request.IdempotencyKey == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonPartialReleaseNeedsKey,
			"a partial release must carry an idempotency key"))
	}
	if vErrs.Count() > 0 {
		return &itStock.ReleaseReservationResult{ClientErrors: *vErrs}, nil
	}

	var result *itStock.ReleaseReservationResult
	err := withReservationTransaction(ctx, func(tranxCtx corectx.Context) error {
		released, err := this.releaseUnderGuard(tranxCtx, request)
		result = released
		return err
	})
	if err != nil {
		return nil, err
	}
	if !result.Refused() && result.ReleasedNow.IsPositive() {
		DrainOutboxNow(ctx)
	}
	return result, nil
}

func (this *StockReservationDomainServiceImpl) releaseUnderGuard(
	ctx corectx.Context, request itStock.ReleaseReservationRequest,
) (*itStock.ReleaseReservationResult, error) {
	vErrs := ft.NewClientErrors()
	reservation, err := loadReservationInOrg(ctx, request.OrgId, request.ReservationId, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ReleaseReservationResult{ClientErrors: *vErrs}, nil
	}

	guardRepo, err := repoFor(models.WarehouseProductGuardSchemaName)
	if err != nil {
		return nil, err
	}
	key := guardKeyOf(*reservation)
	if _, err := LockGuardsForUpdate(ctx, guardRepo.GetBaseRepo(), []GuardKey{key}); err != nil {
		return nil, err
	}
	now, err := DbNowUnderLock(ctx, guardRepo.GetBaseRepo())
	if err != nil {
		return nil, err
	}
	if _, err := MaterializeExpiredInScope(ctx, []GuardKey{key}, now); err != nil {
		return nil, err
	}

	reservation, err = loadReservationInOrg(ctx, request.OrgId, request.ReservationId, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ReleaseReservationResult{ClientErrors: *vErrs}, nil
	}

	// Nothing left to give back: released, consumed, or lapsed and just materialized.
	if derefString(reservation.GetStatus()) != models.StockReservationStatusActive {
		return releaseResultOf(*reservation, decimal.Zero, now), nil
	}

	if request.Quantity.IsPositive() {
		done, err := partialReleaseRecorded(ctx, *reservation, request.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if done {
			return releaseResultOf(*reservation, decimal.Zero, now), nil
		}
	}

	remaining := reservation.RemainingQuantity()
	quantity := remaining
	if request.Quantity.IsPositive() {
		if request.Quantity.GreaterThan(remaining) {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationQuantityExceeded,
				"releasing "+request.Quantity.String()+" exceeds the "+remaining.String()+" still held"))
			return &itStock.ReleaseReservationResult{ClientErrors: *vErrs}, nil
		}
		quantity = request.Quantity
	}

	updated, err := writeRelease(ctx, *reservation, quantity, request.Reason, now)
	if err != nil {
		return nil, err
	}
	if _, err := RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventInventoryReservationReleased,
		AggregateId: derefId(reservation.GetId()),
		OrgId:       string(request.OrgId),
		Payload: map[string]any{
			"reservation_id":     derefId(reservation.GetId()),
			"released_now":       quantity.String(),
			"released_total":     derefDecimal(updated.GetReleasedQuantity()).String(),
			"remaining_quantity": updated.RemainingQuantity().String(),
			"status":             derefString(updated.GetStatus()),
			"reason":             request.Reason,
			"idempotency_key":    request.IdempotencyKey,
		},
		OccurredAt: now.Unix(),
	}); err != nil {
		return nil, err
	}
	return releaseResultOf(*updated, quantity, now), nil
}

// writeRelease raises the released figure and closes the row when nothing remains. released_at
// is now for a deliberate release; expiry writes the deadline instead (see reservation_expiry.go).
func writeRelease(
	ctx corectx.Context, reservation models.StockReservation, quantity decimal.Decimal, reason string, now time.Time,
) (*models.StockReservation, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}
	released := derefDecimal(reservation.GetReleasedQuantity()).Add(quantity)
	status := models.StockReservationStatusActive
	if reservation.RemainingQuantity().Sub(quantity).IsZero() {
		status = models.StockReservationStatusReleased
	}
	releasedAt := model.WrapModelDateTime(now)
	update := dmodel.DynamicFields{
		models.StockReservationFieldId:               derefId(reservation.GetId()),
		models.StockReservationFieldReleasedQuantity: released,
		models.StockReservationFieldStatus:           status,
		models.StockReservationFieldReleasedAt:       releasedAt,
	}
	if reason != "" {
		update[models.StockReservationFieldReleaseReason] = reason
	}
	if _, err := engine.Update(ctx, update); err != nil {
		return nil, errors.Wrap(err, "writeRelease")
	}

	updated := models.NewStockReservationFrom(maps.Clone(reservation.GetFieldData()))
	updated.SetReleasedQuantity(&released)
	updated.SetStatus(&status)
	updated.SetReleasedAt(&releasedAt)
	if reason != "" {
		updated.SetReleaseReason(&reason)
	}
	return updated, nil
}

// partialReleaseRecorded reports whether a partial release under this key was already written,
// by reading the release events of the row: the key travels in the event payload, so no second
// key store is needed and a retry finds its own earlier release.
func partialReleaseRecorded(ctx corectx.Context, reservation models.StockReservation, key string) (bool, error) {
	engine, err := repoFor(models.InventoryIntegrationOutboxSchemaName)
	if err != nil {
		return false, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InventoryOutboxFieldAggregateId, dmodel.Equals, derefId(reservation.GetId())),
		*dmodel.NewSearchNode().NewCondition(models.InventoryOutboxFieldEventType, dmodel.Equals, models.EventInventoryReservationReleased),
	)
	found := false
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		payload, _ := row[models.InventoryOutboxFieldPayload].(map[string]any)
		if recorded, _ := payload["idempotency_key"].(string); recorded == key {
			found = true
		}
	})
	return found, errors.Wrap(err, "partialReleaseRecorded")
}

func releaseResultOf(reservation models.StockReservation, releasedNow decimal.Decimal, now time.Time) *itStock.ReleaseReservationResult {
	return &itStock.ReleaseReservationResult{
		ReservationId:     model.Id(derefId(reservation.GetId())),
		ReleasedNow:       releasedNow,
		ReleasedTotal:     derefDecimal(reservation.GetReleasedQuantity()),
		RemainingQuantity: reservation.RemainingQuantity(),
		Status:            derefString(reservation.GetStatus()),
		EffectiveStatus:   EffectiveStatusOf(reservation, now),
	}
}
