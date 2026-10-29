package services

import (
	"maps"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
)

// Consuming a reservation (CR-INV-SALES-WH-RESERVATION §5.1).
//
// Consumption is the one operation where H and R fall together: the goods physically left, so
// the balance drops and the commitment that covered them is spent. It is recorded as a done stock
// movement from the location the executor actually used, written through the same path every
// other on-hand change takes, and the reservation's consumed figure rises in the same transaction.
// There is deliberately no way to raise consumed_quantity without that movement.

const (
	ReasonReservationNotFound         = "stock_reservation.not_found"
	ReasonReservationExpired          = "stock_reservation.reservation_expired"
	ReasonReservationNotActive        = "stock_reservation.not_active"
	ReasonReservationQuantityExceeded = "stock_reservation.reservation_quantity_exceeded"
	ReasonLocationWarehouseMismatch   = "stock_reservation.location_warehouse_mismatch"
	ReasonActualSourceShort           = "stock_reservation.actual_source_insufficient_stock"
	ReasonNoCustomerLocation          = "stock_reservation.no_customer_location"
	ReasonScopeMismatch               = "stock_reservation.scope_mismatch"
)

// ConsumeReservation records goods that left stock against a reservation.
func (this *StockReservationDomainServiceImpl) ConsumeReservation(
	ctx corectx.Context, request itStock.ConsumeReservationRequest,
) (*itStock.ConsumeReservationResult, error) {
	if vErrs := assertConsumeRequestWellFormed(request); vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	var result *itStock.ConsumeReservationResult
	err := withReservationTransaction(ctx, func(tranxCtx corectx.Context) error {
		consumed, err := this.consumeUnderGuard(tranxCtx, request)
		result = consumed
		return err
	})
	if vErrs, refused := ProtectionViolationOf(err); refused {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}
	if err != nil {
		return nil, err
	}
	if !result.Refused() && !result.Replayed {
		DrainOutboxNow(ctx)
	}
	return result, nil
}

func (this *StockReservationDomainServiceImpl) consumeUnderGuard(
	ctx corectx.Context, request itStock.ConsumeReservationRequest,
) (*itStock.ConsumeReservationResult, error) {
	vErrs := ft.NewClientErrors()

	reservation, err := loadReservationInOrg(ctx, request.OrgId, request.ReservationId, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	// A resent report of the same attempt finds the movement it already made. Checked before the
	// lock: a replay changes nothing, so it need not queue behind anyone.
	if replayed, err := replayedConsumption(ctx, *reservation, request); err != nil || replayed != nil {
		return replayed, err
	}

	key := guardKeyOf(*reservation)
	guardRepo, err := repoFor(models.WarehouseProductGuardSchemaName)
	if err != nil {
		return nil, err
	}
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

	// Re-read under the lock: the row may have been consumed, released or protected since.
	reservation, err = loadReservationInOrg(ctx, request.OrgId, request.ReservationId, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	quantity := totalActualQuantity(request.ActualSources)
	assertReservationCanConsume(*reservation, quantity, now, vErrs)
	if vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	destination, err := resolveConsumeDestination(ctx, request, vErrs)
	if err != nil {
		return nil, err
	}
	if err := assertSourcesInWarehouse(ctx, *reservation, request.ActualSources, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	movementRef, err := issueActualSources(ctx, *reservation, request, destination, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ConsumeReservationResult{ClientErrors: *vErrs}, nil
	}

	updated, err := raiseConsumed(ctx, *reservation, quantity, now)
	if err != nil {
		return nil, err
	}

	eventType := models.EventInventoryReservationPartiallyConsumed
	if derefString(updated.GetStatus()) == models.StockReservationStatusConsumed {
		eventType = models.EventInventoryReservationConsumed
	}
	if _, err := RecordEvent(ctx, RecordEventParams{
		EventType:   eventType,
		AggregateId: string(request.ReservationId),
		OrgId:       string(request.OrgId),
		Payload: map[string]any{
			"reservation_id":       string(request.ReservationId),
			"execution_id":         request.ExecutionId,
			"consumed_now":         quantity.String(),
			"consumed_total":       derefDecimal(updated.GetConsumedQuantity()).String(),
			"remaining_quantity":   updated.RemainingQuantity().String(),
			"inventory_result_ref": movementRef,
		},
		OccurredAt: now.Unix(),
	}); err != nil {
		return nil, err
	}

	return consumeResultOf(*updated, quantity, movementRef, now, false), nil
}

func assertConsumeRequestWellFormed(request itStock.ConsumeReservationRequest) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	refuse := func(field string, message string) {
		vErrs.Append(*ft.NewBusinessViolation(
			field, ft.ErrorKey(ReasonReservationRequestMalformed, "inventory"), message))
	}
	if request.OrgId == "" || request.ReservationId == "" {
		refuse("", "org_id and reservation id are required")
	}
	if request.ExecutionId == "" || request.IdempotencyKey == "" {
		refuse("", "execution_id and idempotency_key are required")
	}
	if len(request.ActualSources) == 0 {
		refuse("actual_sources", "at least one actual source is required: nothing is consumed without a movement")
	}
	for index, source := range request.ActualSources {
		if source.LocationId == "" {
			refuse(requestItemPath("actual_sources", index, "location_id"), "every actual source names a location")
		}
		if !source.Quantity.IsPositive() {
			refuse(requestItemPath("actual_sources", index, "quantity"), "quantity must be greater than zero")
		}
	}
	return vErrs
}

// loadReservationInOrg reads the row and refuses one outside the org the same way as one that does
// not exist, so no caller learns another org's reservation ids.
func loadReservationInOrg(
	ctx corectx.Context, orgId, reservationId model.Id, vErrs *ft.ClientErrors,
) (*models.StockReservation, error) {
	row, err := findRecord(ctx, models.StockReservationSchemaName, models.StockReservationFieldId, string(reservationId))
	if err != nil {
		return nil, err
	}
	if row == nil || stringOf(row, models.StockReservationFieldOrgId) != string(orgId) {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey(ReasonReservationNotFound, "inventory"),
			"reservation '"+string(reservationId)+"' does not exist"))
		return nil, nil
	}
	return models.NewStockReservationFrom(row), nil
}

// assertReservationCanConsume applies the quantity and lifecycle rules under the lock.
func assertReservationCanConsume(
	reservation models.StockReservation, quantity decimal.Decimal, now time.Time, vErrs *ft.ClientErrors,
) {
	switch {
	case derefString(reservation.GetReleaseReason()) == models.StockReservationReleaseReasonExpired,
		derefString(reservation.GetStatus()) == models.StockReservationStatusActive && !IsReservationEffective(reservation, now):
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey(ReasonReservationExpired, "inventory"),
			"the reservation lapsed at "+reservation.GetReservedUntil().GoTime().UTC().Format(time.RFC3339)))
	case derefString(reservation.GetStatus()) != models.StockReservationStatusActive:
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey(ReasonReservationNotActive, "inventory"),
			"the reservation is "+derefString(reservation.GetStatus())+" and holds nothing to consume"))
	case quantity.GreaterThan(reservation.RemainingQuantity()):
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey(ReasonReservationQuantityExceeded, "inventory"),
			"consuming "+quantity.String()+" exceeds the "+reservation.RemainingQuantity().String()+" still held"))
	}
}

func totalActualQuantity(sources []itStock.ActualSource) decimal.Decimal {
	total := decimal.Zero
	for _, source := range sources {
		total = total.Add(source.Quantity)
	}
	return total
}

// resolveConsumeDestination is where the goods went: the caller's location, else the org's
// customer location, which every sale leaves stock to.
func resolveConsumeDestination(
	ctx corectx.Context, request itStock.ConsumeReservationRequest, vErrs *ft.ClientErrors,
) (string, error) {
	if request.DestinationLocationId != "" {
		return string(request.DestinationLocationId), nil
	}
	customer, err := FindLocationByType(ctx, string(request.OrgId), models.InventoryLocationTypeCustomer)
	if err != nil {
		return "", err
	}
	if customer == nil {
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey(ReasonNoCustomerLocation, "inventory"),
			"this organisation has no customer location to issue goods to"))
		return "", nil
	}
	return derefId(customer.GetId()), nil
}

// assertSourcesInWarehouse refuses a source location outside the reservation's warehouse: a hold
// at one warehouse cannot be spent from another's shelves, whatever the executor reports.
func assertSourcesInWarehouse(
	ctx corectx.Context, reservation models.StockReservation, sources []itStock.ActualSource, vErrs *ft.ClientErrors,
) error {
	for index, source := range sources {
		location, err := findRecord(ctx, models.InventoryLocationSchemaName, models.InventoryLocationFieldId, string(source.LocationId))
		if err != nil {
			return err
		}
		warehouseId, err := warehouseOfLocation(ctx, location)
		if err != nil {
			return err
		}
		if location == nil ||
			stringOf(location, models.InventoryLocationFieldOrgId) != derefId(reservation.GetOrgId()) ||
			warehouseId != derefId(reservation.GetWarehouseId()) {
			vErrs.Append(*ft.NewBusinessViolation(requestItemPath("actual_sources", index, "location_id"),
				ft.ErrorKey(ReasonLocationWarehouseMismatch, "inventory"),
				"location '"+string(source.LocationId)+"' does not belong to the reservation's warehouse"))
		}
	}
	return nil
}

func warehouseOfLocation(ctx corectx.Context, location dmodel.DynamicFields) (string, error) {
	seen := map[string]bool{}
	for location != nil {
		if warehouseId := stringOf(location, models.InventoryLocationFieldWarehouseId); warehouseId != "" {
			return warehouseId, nil
		}
		parentId := stringOf(location, models.InventoryLocationFieldParentLocationId)
		if parentId == "" || seen[parentId] {
			return "", nil
		}
		seen[parentId] = true

		parent, err := findRecord(ctx, models.InventoryLocationSchemaName, models.InventoryLocationFieldId, parentId)
		if err != nil {
			return "", err
		}
		location = parent
	}
	return "", nil
}

// issueActualSources writes one done movement per actual source, from the location the executor
// used to the destination. Each source's quants are locked and checked first: the goods must be
// there, and not already claimed by a transfer hold on that same balance.
func issueActualSources(
	ctx corectx.Context, reservation models.StockReservation, request itStock.ConsumeReservationRequest,
	destination string, vErrs *ft.ClientErrors,
) (string, error) {
	quantRepo, err := repoFor(models.StockQuantSchemaName)
	if err != nil {
		return "", err
	}
	operationType, operationCode := "", ""
	if request.OperationTypeId != "" {
		operationType = string(request.OperationTypeId)
		operationCode = models.StockOperationCodeOutgoing
	}

	type plannedSource struct {
		source         itStock.ActualSource
		idempotencyKey string
		pieces         []Allocation
	}
	planned := make([]plannedSource, 0, len(request.ActualSources))
	refusedBefore := vErrs.Count()
	for index, source := range request.ActualSources {
		locked, err := LockQuantsForUpdate(ctx, quantRepo.GetBaseRepo(), QuantLockKey{
			OrgId:            model.Id(derefId(reservation.GetOrgId())),
			ProductVariantId: model.Id(derefId(reservation.GetProductVariantId())),
			LocationId:       source.LocationId,
		})
		if err != nil {
			return "", err
		}
		if available := availableForSource(locked, source); available.LessThan(source.Quantity) {
			vErrs.Append(*ft.NewAnonymousBusinessViolation(
				ft.ErrorKey(ReasonActualSourceShort, "inventory"),
				"location '"+string(source.LocationId)+"' holds "+available.String()+" free, not the "+
					source.Quantity.String()+" reported as taken"))
			continue
		}

		idempotencyKey := request.IdempotencyKey
		if index > 0 {
			idempotencyKey = request.IdempotencyKey + ":" + string(source.LocationId)
		}
		planned = append(planned, plannedSource{
			source:         source,
			idempotencyKey: idempotencyKey,
			pieces:         piecesOfSource(locked, source),
		})
	}
	if vErrs.Count() > refusedBefore {
		return "", nil
	}

	var firstRef string
	for _, plan := range planned {
		for index, piece := range plan.pieces {
			idempotencyKey := plan.idempotencyKey
			if index > 0 {
				idempotencyKey = plan.idempotencyKey + ":" + strconv.Itoa(index)
			}
			result, cErrs, err := ApplyCorrectionMovement(ctx, CorrectionRequest{
				OrgId:                 derefId(reservation.GetOrgId()),
				ProductVariantId:      derefId(reservation.GetProductVariantId()),
				Quantity:              piece.Quantity,
				SourceLocationId:      string(plan.source.LocationId),
				DestinationLocationId: destination,
				LotRef:                piece.LotRef,
				PackageRef:            piece.PackageRef,
				OwnerRef:              piece.OwnerRef,
				OriginReference:       request.OriginReference,
				Note:                  "reservation " + derefId(reservation.GetId()) + " execution " + request.ExecutionId,
				OperationTypeId:       operationType,
				OperationCode:         operationCode,
				IdempotencyKey:        idempotencyKey,
			})
			if err != nil {
				return "", err
			}
			if cErrs != nil && cErrs.Count() > 0 {
				vErrs.ConcatPtr(cErrs)
				continue
			}
			if firstRef == "" {
				firstRef = result.TransferId
			}
		}
	}
	return firstRef, nil
}

func namesDimension(source itStock.ActualSource) bool {
	return source.LotRef != "" || source.PackageRef != "" || source.OwnerRef != ""
}

func availableForSource(locked []LockedQuant, source itStock.ActualSource) decimal.Decimal {
	if namesDimension(source) {
		return availableInDimension(locked, source)
	}
	available := decimal.Zero
	for _, quant := range locked {
		if quant.Available().IsPositive() {
			available = available.Add(quant.Available())
		}
	}
	return available
}

func piecesOfSource(locked []LockedQuant, source itStock.ActualSource) []Allocation {
	if namesDimension(source) {
		return []Allocation{{
			Quantity:   source.Quantity,
			LotRef:     source.LotRef,
			PackageRef: source.PackageRef,
			OwnerRef:   source.OwnerRef,
		}}
	}
	allocations, _ := AllocateFromQuants(source.Quantity, locked)
	return allocations
}

// availableInDimension sums what the locked quants of exactly this lot/package/owner can still
// give: on-hand less what transfer holds already claim on them.
func availableInDimension(locked []LockedQuant, source itStock.ActualSource) decimal.Decimal {
	available := decimal.Zero
	for _, quant := range locked {
		if quant.LotRef != source.LotRef || quant.PackageRef != source.PackageRef || quant.OwnerRef != source.OwnerRef {
			continue
		}
		available = available.Add(quant.Available())
	}
	return available
}

// raiseConsumed adds the consumed quantity to the row and closes it when nothing remains. Written
// through the repository: the schema keeps these columns closed to clients.
func raiseConsumed(
	ctx corectx.Context, reservation models.StockReservation, quantity decimal.Decimal, now time.Time,
) (*models.StockReservation, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}
	consumed := derefDecimal(reservation.GetConsumedQuantity()).Add(quantity)
	update := dmodel.DynamicFields{
		models.StockReservationFieldId:               derefId(reservation.GetId()),
		models.StockReservationFieldConsumedQuantity: consumed,
	}
	status := models.StockReservationStatusActive
	if reservation.RemainingQuantity().Sub(quantity).IsZero() {
		status = models.StockReservationStatusConsumed
		update[models.StockReservationFieldStatus] = status
	}
	if _, err := engine.Update(ctx, update); err != nil {
		return nil, errors.Wrap(err, "raiseConsumed")
	}

	updated := models.NewStockReservationFrom(maps.Clone(reservation.GetFieldData()))
	updated.SetConsumedQuantity(&consumed)
	updated.SetStatus(&status)
	return updated, nil
}

// replayedConsumption answers a resent report from the movement it already made: a transfer
// carrying this request's idempotency key for this org. A key on an in-flight document means
// the earlier attempt did not complete, so the retry proceeds.
func replayedConsumption(
	ctx corectx.Context, reservation models.StockReservation, request itStock.ConsumeReservationRequest,
) (*itStock.ConsumeReservationResult, error) {
	engine, err := repoFor(models.StockTransferSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockTransferFieldOrgId, dmodel.Equals, string(request.OrgId)),
		*dmodel.NewSearchNode().NewCondition(models.StockTransferFieldIdempotencyKey, dmodel.Equals, request.IdempotencyKey),
		*dmodel.NewSearchNode().NewCondition(models.StockTransferFieldStatus, dmodel.Equals, models.StockTransferStatusDone),
	)
	found, err := engine.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: 1})
	if err != nil {
		return nil, errors.Wrap(err, "replayedConsumption")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	transfer := models.NewStockTransferFrom(found.Data.Items[0])
	return consumeResultOf(reservation, decimal.Zero, derefString(transfer.GetId()), time.Now().UTC(), true), nil
}

func consumeResultOf(
	reservation models.StockReservation, consumedNow decimal.Decimal, movementRef string, now time.Time, replayed bool,
) *itStock.ConsumeReservationResult {
	return &itStock.ConsumeReservationResult{
		ReservationId:      model.Id(derefId(reservation.GetId())),
		ConsumedNow:        consumedNow,
		ConsumedTotal:      derefDecimal(reservation.GetConsumedQuantity()),
		RemainingQuantity:  reservation.RemainingQuantity(),
		Status:             derefString(reservation.GetStatus()),
		EffectiveStatus:    EffectiveStatusOf(reservation, now),
		InventoryResultRef: movementRef,
		Replayed:           replayed,
	}
}
