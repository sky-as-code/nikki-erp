package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Protecting a paid demand's reservations (CR-INV-SALES-WH-RESERVATION §7.2).
//
// Once the order behind a hold is paid, the deadline that protected the shop from an abandoned
// order protects nobody: the goods now belong to a customer until they are handed over or the
// order is cancelled. Protection clears reserved_until on every reservation of the demand that is
// still in force. One that lapsed before the lock was taken is NOT revived — the stock may already
// be somebody else's — it is materialized as expired and reported back, and the caller decides
// whether to reserve afresh. This is reachable only through the port: a client cannot declare its
// own order paid.

// ProtectPaidReservations clears the deadline on the still-effective reservations of a source.
func (this *StockReservationDomainServiceImpl) ProtectPaidReservations(
	ctx corectx.Context, request itStock.ProtectPaidReservationsRequest,
) (*itStock.ProtectPaidReservationsResult, error) {
	vErrs := ft.NewClientErrors()
	if request.OrgId == "" || request.SourceModule == "" || request.SourceType == "" || request.SourceId == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
			"org_id, source_module, source_type and source_id are required"))
		return &itStock.ProtectPaidReservationsResult{ClientErrors: *vErrs}, nil
	}
	if request.SourceRevision <= 0 {
		request.SourceRevision = 1
	}

	var result *itStock.ProtectPaidReservationsResult
	err := withReservationTransaction(ctx, func(tranxCtx corectx.Context) error {
		protected, err := protectUnderGuard(tranxCtx, request)
		result = protected
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(result.Protected) > 0 || len(result.Expired) > 0 {
		DrainOutboxNow(ctx)
	}
	return result, nil
}

func protectUnderGuard(
	ctx corectx.Context, request itStock.ProtectPaidReservationsRequest,
) (*itStock.ProtectPaidReservationsResult, error) {
	reservations, err := reservationsOfSource(ctx, request.OrgId, request.SourceModule, request.SourceType,
		request.SourceId, request.SourceRevision)
	if err != nil {
		return nil, err
	}
	result := &itStock.ProtectPaidReservationsResult{}
	if len(reservations) == 0 {
		return result, nil
	}

	keys := make([]GuardKey, 0, len(reservations))
	for _, reservation := range reservations {
		keys = append(keys, guardKeyOf(reservation))
	}
	guardRepo, err := repoFor(models.WarehouseProductGuardSchemaName)
	if err != nil {
		return nil, err
	}
	if _, err := LockGuardsForUpdate(ctx, guardRepo.GetBaseRepo(), keys); err != nil {
		return nil, err
	}
	now, err := DbNowUnderLock(ctx, guardRepo.GetBaseRepo())
	if err != nil {
		return nil, err
	}
	if _, err := MaterializeExpiredInScope(ctx, keys, now); err != nil {
		return nil, err
	}

	// Re-read under the lock: expiry may just have been materialized, or a concurrent consume
	// may have closed a row.
	reservations, err = reservationsOfSource(ctx, request.OrgId, request.SourceModule, request.SourceType,
		request.SourceId, request.SourceRevision)
	if err != nil {
		return nil, err
	}

	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}
	for _, reservation := range reservations {
		id := model.Id(derefId(reservation.GetId()))
		switch {
		case derefString(reservation.GetReleaseReason()) == models.StockReservationReleaseReasonExpired:
			result.Expired = append(result.Expired, id)
		case derefString(reservation.GetStatus()) != models.StockReservationStatusActive:
			// Consumed or deliberately released: nothing to protect, nothing to report.
		case reservation.GetReservedUntil() == nil:
			result.AlreadyProtected = append(result.AlreadyProtected, id)
		default:
			previous := reservation.GetReservedUntil().GoTime().UTC()
			if _, err := engine.Update(ctx, dmodel.DynamicFields{
				models.StockReservationFieldId:            string(id),
				models.StockReservationFieldReservedUntil: nil,
			}); err != nil {
				return nil, errors.Wrap(err, "ProtectPaidReservations")
			}
			if _, err := RecordEvent(ctx, RecordEventParams{
				EventType:   models.EventInventoryReservationPaidProtected,
				AggregateId: string(id),
				OrgId:       string(request.OrgId),
				Payload: map[string]any{
					"reservation_id":          string(id),
					"previous_reserved_until": previous.Format(time.RFC3339),
					"payment_reference":       request.PaymentReference,
					"idempotency_key":         request.IdempotencyKey,
				},
				OccurredAt: now.Unix(),
			}); err != nil {
				return nil, err
			}
			result.Protected = append(result.Protected, id)
		}
	}
	return result, nil
}
