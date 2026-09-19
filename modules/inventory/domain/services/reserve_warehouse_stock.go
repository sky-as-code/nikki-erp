package services

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Reserving at warehouse level (CR-INV-SALES-WH-RESERVATION §4.4).
//
// The five steps are fixed: validate, normalise to base units, lock every scope's guard, read the
// database clock and recompute availability under the lock, then write every line or none. The
// availability figure that decides the reserve is the one computed after the guard was taken; a
// figure read before queuing is stale by however long the wait was.

var _ itStock.WarehouseReservationService = (*StockReservationDomainServiceImpl)(nil)

// Refusal keys. Lowercase dotted per the module convention; the CR names them in upper case.
const (
	ReasonReservationRequestMalformed    = "stock_reservation.request_malformed"
	ReasonWarehouseNotFound              = "stock_reservation.warehouse_not_found"
	ReasonWarehouseSuspended             = "stock_reservation.warehouse_suspended"
	ReasonResourceArchived               = "stock_reservation.resource_archived"
	ReasonVariantNotFound                = "stock_reservation.variant_not_found"
	ReasonVariantNotStockable            = "stock_reservation.not_stockable"
	ReasonDeadlineNotInFuture            = "stock_reservation.deadline_not_in_future"
	ReasonInsufficientWarehouseStock     = "stock_reservation.insufficient_warehouse_stock"
	ReasonReservationIdempotencyConflict = "stock_reservation.idempotency_conflict"
)

// ReserveWarehouseStock holds every line of the request at the warehouse, or nothing.
func (this *StockReservationDomainServiceImpl) ReserveWarehouseStock(
	ctx corectx.Context, request itStock.ReserveWarehouseStockRequest,
) (*itStock.ReserveWarehouseStockResult, error) {
	if request.SourceRevision <= 0 {
		request.SourceRevision = 1
	}
	if vErrs := assertReserveRequestWellFormed(request); vErrs.Count() > 0 {
		return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}, nil
	}

	var result *itStock.ReserveWarehouseStockResult
	err := withReservationTransaction(ctx, func(tranxCtx corectx.Context) error {
		reserved, err := this.reserveUnderGuard(tranxCtx, request)
		result = reserved
		return err
	})
	if err != nil {
		return nil, err
	}
	if !result.Refused() && !result.Replayed {
		DrainOutboxNow(ctx)
	}
	return result, nil
}

func (this *StockReservationDomainServiceImpl) reserveUnderGuard(
	ctx corectx.Context, request itStock.ReserveWarehouseStockRequest,
) (*itStock.ReserveWarehouseStockResult, error) {
	vErrs := ft.NewClientErrors()

	if err := assertWarehouseReservable(ctx, request.WarehouseId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}, nil
	}

	lines, err := this.normaliseReservationLines(ctx, request, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}, nil
	}

	fingerprint := reservationFingerprint(request, lines)

	// Guards first, then the replay check: the check reads rows a concurrent first attempt may
	// still be writing, and only the lock makes the answer final.
	keys := make([]GuardKey, 0, len(lines))
	for _, line := range lines {
		keys = append(keys, GuardKey{
			OrgId: request.OrgId, WarehouseId: request.WarehouseId, ProductVariantId: line.ProductVariantId,
		})
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

	existing, err := reservationsOfSource(ctx, request.OrgId, request.SourceModule, request.SourceType,
		request.SourceId, request.SourceRevision)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return replayReservation(existing, request, fingerprint, now), nil
	}

	if request.ReservedUntil != nil && !request.ReservedUntil.After(now) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonDeadlineNotInFuture,
			"reserved_until must lie in the future"))
		return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}, nil
	}

	if err := assertScopesCanCover(ctx, request, lines, now, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}, nil
	}

	written, err := writeReservations(ctx, request, lines, fingerprint, now)
	if err != nil {
		return nil, err
	}
	return &itStock.ReserveWarehouseStockResult{Reservations: written}, nil
}

// normalisedLine is a request line expressed in the variant's base unit.
type normalisedLine struct {
	SourceLineId     string
	ProductVariantId model.Id
	BaseUomId        model.Id
	Quantity         decimal.Decimal
}

func assertReserveRequestWellFormed(request itStock.ReserveWarehouseStockRequest) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	refuse := func(message string) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed, message))
	}
	if request.OrgId == "" {
		refuse("org_id is required")
	}
	if request.WarehouseId == "" {
		refuse("warehouse_id is required")
	}
	if request.SourceModule == "" || request.SourceType == "" || request.SourceId == "" {
		refuse("source_module, source_type and source_id are required")
	}
	if request.IdempotencyKey == "" {
		refuse("idempotency_key is required")
	}
	if len(request.Lines) == 0 {
		refuse("at least one line is required")
	}
	seenLines := map[string]bool{}
	for index, line := range request.Lines {
		if line.ProductVariantId == "" {
			refuse("line " + strconv.Itoa(index+1) + ": product_variant_id is required")
		}
		if !line.Quantity.IsPositive() {
			refuse("line " + strconv.Itoa(index+1) + ": quantity must be greater than zero")
		}
		if seenLines[line.SourceLineId] {
			refuse("line " + strconv.Itoa(index+1) + ": source_line_id '" + line.SourceLineId + "' appears twice")
		}
		seenLines[line.SourceLineId] = true
	}
	return vErrs
}

// assertWarehouseReservable refuses a warehouse that cannot take a new commitment. A suspended
// warehouse keeps the commitments it has; it just takes no more.
func assertWarehouseReservable(ctx corectx.Context, warehouseId model.Id, vErrs *ft.ClientErrors) error {
	engine, err := repoFor(models.WarehouseSchemaName)
	if err != nil {
		return err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{models.WarehouseFieldId: string(warehouseId)})
	if err != nil {
		return errors.Wrap(err, "assertWarehouseReservable")
	}
	if found == nil || !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonWarehouseNotFound,
			"warehouse '"+string(warehouseId)+"' does not exist"))
		return nil
	}
	warehouse := models.NewWarehouseFrom(found.Data)
	if derefBool(warehouse.GetIsArchived()) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonResourceArchived,
			"warehouse '"+string(warehouseId)+"' is archived"))
		return nil
	}
	if derefString(warehouse.GetStatus()) != models.WarehouseStatusActive {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonWarehouseSuspended,
			"warehouse '"+string(warehouseId)+"' is suspended"))
	}
	return nil
}

// normaliseReservationLines resolves each line's variant, refuses one that cannot hold stock, and
// converts the quantity into the variant's base unit.
func (this *StockReservationDomainServiceImpl) normaliseReservationLines(
	ctx corectx.Context, request itStock.ReserveWarehouseStockRequest, vErrs *ft.ClientErrors,
) ([]normalisedLine, error) {
	lines := make([]normalisedLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		baseUomId, err := this.resolveReservableVariant(ctx, request.OrgId, line.ProductVariantId, vErrs)
		if err != nil {
			return nil, err
		}
		if baseUomId == "" {
			continue
		}

		quantity := line.Quantity
		if line.UomId != "" && line.UomId != baseUomId {
			converted, err := this.uom.Convert(ctx, itExt.ConvertQuantityQuery{
				Quantity: line.Quantity, SourceUomId: line.UomId, TargetUomId: baseUomId,
			})
			if err != nil {
				return nil, errors.Wrap(err, "converting a reservation quantity to the base unit")
			}
			if converted == nil || converted.ClientErrors.Count() > 0 || !converted.HasData {
				vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
					"line '"+line.SourceLineId+"': quantity cannot be converted from unit '"+string(line.UomId)+"' to the base unit"))
				continue
			}
			quantity = converted.Data.Quantity
		}
		lines = append(lines, normalisedLine{
			SourceLineId:     line.SourceLineId,
			ProductVariantId: line.ProductVariantId,
			BaseUomId:        baseUomId,
			Quantity:         quantity,
		})
	}
	return lines, nil
}

// resolveReservableVariant returns the variant's base unit, or empty after appending why it cannot
// be reserved: a service or combo product holds no stock, so no reservation can be made for it.
// The base unit is the template's stock configuration when one exists, else the template's own unit.
func (this *StockReservationDomainServiceImpl) resolveReservableVariant(
	ctx corectx.Context, orgId, variantId model.Id, vErrs *ft.ClientErrors,
) (model.Id, error) {
	variant, err := findRecord(ctx, models.ProductVariantSchemaName, models.ProductVariantFieldId, string(variantId))
	if err != nil {
		return "", err
	}
	if variant == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonVariantNotFound,
			"product variant '"+string(variantId)+"' does not exist"))
		return "", nil
	}
	templateId := stringOf(variant, models.ProductVariantFieldProductTemplateId)
	template, err := findRecord(ctx, models.ProductTemplateSchemaName, models.ProductTemplateFieldId, templateId)
	if err != nil {
		return "", err
	}
	if template == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonVariantNotFound,
			"product variant '"+string(variantId)+"' has no template"))
		return "", nil
	}

	productType, err := findRecord(ctx, models.ProductTypeSchemaName, models.ProductTypeFieldId,
		stringOf(template, models.ProductTemplateFieldProductTypeId))
	if err != nil {
		return "", err
	}
	if productType == nil || !derefBool(models.NewProductTypeFrom(productType).GetSupportsStock()) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonVariantNotStockable,
			"product variant '"+string(variantId)+"' is not a stockable product"))
		return "", nil
	}

	config, err := findRecordWhere(ctx, models.StockProductConfigSchemaName, dmodel.DynamicFields{
		models.StockProductConfigFieldProductTemplateId: templateId,
		models.StockProductConfigFieldOrgId:             string(orgId),
	})
	if err != nil {
		return "", err
	}
	if config != nil {
		if uomId := stringOf(config, models.StockProductConfigFieldInventoryUomId); uomId != "" {
			return model.Id(uomId), nil
		}
	}
	uomId := stringOf(template, models.ProductTemplateFieldUomId)
	if uomId == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonVariantNotStockable,
			"product variant '"+string(variantId)+"' has no unit of measure"))
		return "", nil
	}
	return model.Id(uomId), nil
}

// assertScopesCanCover checks every variant's total against availability under the held guard.
// All-or-nothing: one short variant refuses the whole request, naming every shortage so the
// caller learns them in one round trip.
func assertScopesCanCover(
	ctx corectx.Context, request itStock.ReserveWarehouseStockRequest, lines []normalisedLine,
	now time.Time, vErrs *ft.ClientErrors,
) error {
	totals := map[model.Id]decimal.Decimal{}
	order := make([]model.Id, 0, len(lines))
	for _, line := range lines {
		if _, seen := totals[line.ProductVariantId]; !seen {
			order = append(order, line.ProductVariantId)
		}
		totals[line.ProductVariantId] = totals[line.ProductVariantId].Add(line.Quantity)
	}

	for _, variantId := range order {
		availability, err := ComputeWarehouseAvailability(ctx, GuardKey{
			OrgId: request.OrgId, WarehouseId: request.WarehouseId, ProductVariantId: variantId,
		}, now)
		if err != nil {
			return err
		}
		if shortage := availability.Shortage(totals[variantId]); shortage.IsPositive() {
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonInsufficientWarehouseStock,
				"product variant '"+string(variantId)+"' is short by "+shortage.String()+
					" (requested "+totals[variantId].String()+", available "+availability.Available.String()+")",
				map[string]any{
					"product_variant_id": string(variantId),
					"requested":          totals[variantId].String(),
					"available":          availability.Available.String(),
					"shortage":           shortage.String(),
				},
			))
		}
	}
	return nil
}

// writeReservations inserts one row per line and announces each, inside the caller's transaction.
func writeReservations(
	ctx corectx.Context, request itStock.ReserveWarehouseStockRequest, lines []normalisedLine,
	fingerprint string, now time.Time,
) ([]itStock.ReservedLine, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}

	written := make([]itStock.ReservedLine, 0, len(lines))
	for _, line := range lines {
		id, err := model.NewId()
		if err != nil {
			return nil, err
		}
		record := dmodel.DynamicFields{
			models.StockReservationFieldId:                 string(*id),
			models.StockReservationFieldOrgId:              string(request.OrgId),
			models.StockReservationFieldWarehouseId:        string(request.WarehouseId),
			models.StockReservationFieldProductVariantId:   string(line.ProductVariantId),
			models.StockReservationFieldBaseUomId:          string(line.BaseUomId),
			models.StockReservationFieldQuantity:           line.Quantity,
			models.StockReservationFieldConsumedQuantity:   decimal.Zero,
			models.StockReservationFieldReleasedQuantity:   decimal.Zero,
			models.StockReservationFieldStatus:             models.StockReservationStatusActive,
			models.StockReservationFieldSourceModule:       request.SourceModule,
			models.StockReservationFieldSourceType:         request.SourceType,
			models.StockReservationFieldSourceId:           request.SourceId,
			models.StockReservationFieldSourceLineId:       line.SourceLineId,
			models.StockReservationFieldSourceRevision:     request.SourceRevision,
			models.StockReservationFieldIdempotencyKey:     request.IdempotencyKey,
			models.StockReservationFieldRequestFingerprint: fingerprint,
		}
		if request.ReservedUntil != nil {
			record[models.StockReservationFieldReservedUntil] = model.WrapModelDateTime(*request.ReservedUntil)
		}
		if _, err := engine.Insert(ctx, record); err != nil {
			return nil, errors.Wrap(err, "writing a stock reservation")
		}

		payload := map[string]any{
			"reservation_id":     string(*id),
			"warehouse_id":       string(request.WarehouseId),
			"product_variant_id": string(line.ProductVariantId),
			"base_uom_id":        string(line.BaseUomId),
			"quantity":           line.Quantity.String(),
			"source_module":      request.SourceModule,
			"source_type":        request.SourceType,
			"source_id":          request.SourceId,
			"source_line_id":     line.SourceLineId,
			"source_revision":    request.SourceRevision,
		}
		if request.ReservedUntil != nil {
			payload["reserved_until"] = request.ReservedUntil.UTC().Format(time.RFC3339)
		}
		if _, err := RecordEvent(ctx, RecordEventParams{
			EventType:   models.EventInventoryReservationCreated,
			AggregateId: string(*id),
			OrgId:       string(request.OrgId),
			Payload:     payload,
			OccurredAt:  now.Unix(),
		}); err != nil {
			return nil, err
		}

		written = append(written, itStock.ReservedLine{
			ReservationId:     *id,
			SourceLineId:      line.SourceLineId,
			ProductVariantId:  line.ProductVariantId,
			BaseUomId:         line.BaseUomId,
			Quantity:          line.Quantity,
			RemainingQuantity: line.Quantity,
			Status:            models.StockReservationStatusActive,
			EffectiveStatus:   models.StockReservationStatusActive,
			ReservedUntil:     request.ReservedUntil,
		})
	}
	return written, nil
}

// replayReservation answers a repeated request from the stored rows. The same key with the same
// payload is a retry and gets what exists, as it stands now; anything else wearing the same
// source is a conflict, because the caller must use a new source revision for a new need.
func replayReservation(
	existing []models.StockReservation, request itStock.ReserveWarehouseStockRequest,
	fingerprint string, now time.Time,
) *itStock.ReserveWarehouseStockResult {
	for _, reservation := range existing {
		if derefString(reservation.GetIdempotencyKey()) != request.IdempotencyKey ||
			derefString(reservation.GetRequestFingerprint()) != fingerprint {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationIdempotencyConflict,
				"this demand was already reserved under a different request; use a new source revision"))
			return &itStock.ReserveWarehouseStockResult{ClientErrors: *vErrs}
		}
	}
	lines := make([]itStock.ReservedLine, 0, len(existing))
	for _, reservation := range existing {
		lines = append(lines, reservedLineOf(reservation, now))
	}
	return &itStock.ReserveWarehouseStockResult{Reservations: lines, Replayed: true}
}

// reservedLineOf projects a stored row as the caller sees it, with the effective status decided
// on the clock passed in.
func reservedLineOf(reservation models.StockReservation, now time.Time) itStock.ReservedLine {
	line := itStock.ReservedLine{
		ReservationId:     model.Id(derefId(reservation.GetId())),
		SourceLineId:      derefString(reservation.GetSourceLineId()),
		ProductVariantId:  model.Id(derefId(reservation.GetProductVariantId())),
		BaseUomId:         model.Id(derefId(reservation.GetBaseUomId())),
		Quantity:          derefDecimal(reservation.GetQuantity()),
		RemainingQuantity: reservation.RemainingQuantity(),
		Status:            derefString(reservation.GetStatus()),
		EffectiveStatus:   EffectiveStatusOf(reservation, now),
	}
	if until := reservation.GetReservedUntil(); until != nil {
		goTime := until.GoTime()
		line.ReservedUntil = &goTime
	}
	return line
}

// EffectiveStatusOf is the status a reader should act on, decided on the given clock.
func EffectiveStatusOf(reservation models.StockReservation, now time.Time) string {
	status := derefString(reservation.GetStatus())
	if status == models.StockReservationStatusActive && !IsReservationEffective(reservation, now) {
		return models.StockReservationEffectiveStatusExpired
	}
	return status
}

// reservationFingerprint hashes the business payload of a reserve request, so a retry can be told
// apart from a different request under the same key. Lines are sorted so ordering is not payload.
func reservationFingerprint(request itStock.ReserveWarehouseStockRequest, lines []normalisedLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		parts = append(parts, strings.Join([]string{
			line.SourceLineId, string(line.ProductVariantId), string(line.BaseUomId), line.Quantity.String(),
		}, "\x1f"))
	}
	sort.Strings(parts)

	deadline := ""
	if request.ReservedUntil != nil {
		deadline = request.ReservedUntil.UTC().Format(time.RFC3339Nano)
	}
	head := strings.Join([]string{
		string(request.OrgId), string(request.WarehouseId), deadline,
		request.SourceModule, request.SourceType, request.SourceId, strconv.Itoa(int(request.SourceRevision)),
	}, "\x1f")

	digest := sha256.Sum256([]byte(strings.Join(append([]string{head}, parts...), "\x1e")))
	return hex.EncodeToString(digest[:])
}

// reservationsOfSource lists every reservation row written for one demand revision, in any
// stored status.
func reservationsOfSource(
	ctx corectx.Context, orgId model.Id, sourceModule, sourceType, sourceId string, revision int32,
) ([]models.StockReservation, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldOrgId, dmodel.Equals, string(orgId)),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldSourceModule, dmodel.Equals, sourceModule),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldSourceType, dmodel.Equals, sourceType),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldSourceId, dmodel.Equals, sourceId),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldSourceRevision, dmodel.Equals, revision),
	)
	reservations := make([]models.StockReservation, 0, 4)
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		reservations = append(reservations, *models.NewStockReservationFrom(row))
	})
	return reservations, errors.Wrap(err, "reservationsOfSource")
}

// withReservationTransaction runs body inside the caller's transaction when one is ambient, else
// inside one of its own. Joining is what lets a selling module reserve, confirm and bill under one
// commit when it shares the database.
func withReservationTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	if ctx != nil && ctx.GetDbTranx() != nil {
		return body(ctx)
	}

	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return err
	}
	tranx, err := engine.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withReservationTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withReservationTransaction")
}

// findRecord reads one row by a single key, or nil when there is none.
func findRecord(ctx corectx.Context, schemaName, field, value string) (dmodel.DynamicFields, error) {
	if value == "" {
		return nil, nil
	}
	return findRecordWhere(ctx, schemaName, dmodel.DynamicFields{field: value})
}

func findRecordWhere(ctx corectx.Context, schemaName string, keys dmodel.DynamicFields) (dmodel.DynamicFields, error) {
	engine, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.FindByKeys(ctx, keys)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", schemaName)
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}
