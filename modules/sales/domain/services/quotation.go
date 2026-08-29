package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Quotations and their conversion to orders.
//
// The conversion re-prices rather than copying the quoted numbers: a copied total has no adjustment
// chain, so the price explanation could not rebuild it, and a six-month-old quotation would create an
// order at prices, taxes and promotions that no longer exist. The quotation carries the lines, not
// the money; valid_until keeps that honest by refusing an offer outside its window rather than
// silently repricing it. Holding a stale price deliberately is a manual discount on the resulting
// order, which is permission-gated and audited.
//
// Conversion is idempotent through converted_sales_order_id: a second accept returns the first
// accept's order, since two orders from one acceptance is two deliveries and two invoices. The column
// is set in the same transaction as the status move, so the two cannot disagree.

// The refusal reasons the quotation operations can produce.
const (
	ReasonQuotationNotFound       = "sales_quotation.not_found"
	ReasonQuotationNotConvertible = "sales_quotation.not_convertible"
	ReasonQuotationExpired        = "sales_quotation.expired"
	ReasonQuotationHasNoLines     = "sales_quotation.no_lines"
	ReasonQuotationBadTransition  = "sales_quotation.invalid_transition"
)

// ConvertQuotationParams is what turning an offer into a sale needs.
type ConvertQuotationParams struct {
	SalesQuotationId string

	// SalesPointId names where the sale will be made. Supplied at conversion rather than taken from
	// the quotation, because a quotation may be prepared centrally with no point decided, and the
	// order's anti-spoofing rule derives the channel from the point, which must be current.
	SalesPointId string

	// IdempotencyKey is passed through to the order: converted_sales_order_id stops a second accept,
	// and this stops a retry of the first one racing itself.
	IdempotencyKey string
}

// ConvertQuotationResult is what the conversion produced.
type ConvertQuotationResult struct {
	SalesQuotationId string
	SalesOrderId     string
	OrderNumber      string

	// AlreadyConverted marks the replay path: the quotation had already become the order returned here.
	AlreadyConverted bool

	// Both totals are returned so a caller can see whether repricing moved the number, before handing
	// an order to a customer who is holding the quotation.
	QuotedTotal decimal.Decimal
	OrderTotal  decimal.Decimal
}

// ConvertQuotation turns an accepted offer into a sales order.
func ConvertQuotation(
	ctx corectx.Context,
	params ConvertQuotationParams,
	taxSvc itExt.TaxCalculationExtService,
	products itExt.ProductVariantExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
) (*ConvertQuotationResult, *ft.ClientErrors, error) {
	quotation, err := loadRecord(ctx,
		models.SalesQuotationSchemaName, models.SalesQuotationFieldId, params.SalesQuotationId)
	if err != nil {
		return nil, nil, err
	}
	if quotation == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_quotation_id", ReasonQuotationNotFound,
			"no quotation exists with id '"+params.SalesQuotationId+"'"))
		return nil, vErrs, nil
	}

	// The replay check comes before the gates: a retry of a conversion that already happened must
	// succeed even if the quotation has since expired, or the caller retries against an order that
	// already exists.
	typed := models.NewSalesQuotationFrom(quotation)
	if typed.IsConverted() {
		return replayConversion(ctx, quotation)
	}

	if vErrs := assertConvertible(quotation); vErrs != nil {
		return nil, vErrs, nil
	}

	lines, err := searchBy(ctx, models.SalesQuotationLineSchemaName,
		models.SalesQuotationLineFieldQuotationId, params.SalesQuotationId)
	if err != nil {
		return nil, nil, err
	}
	if len(lines) == 0 {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_quotation_id", ReasonQuotationHasNoLines,
			"a quotation with no lines offers nothing to convert"))
		return nil, vErrs, nil
	}

	// Created through the ordinary path, so the anti-spoofing rule, idempotency handling and repricing
	// are identical to a directly-created order.
	created, vErrs, err := CreateOrder(ctx, CreateOrderParams{
		SalesPointId:      params.SalesPointId,
		CustomerReference: stringOf(quotation, models.SalesQuotationFieldCustomerRef),
		CurrencyCode:      stringOf(quotation, models.SalesQuotationFieldCurrencyCode),
		Lines:             orderLinesFromQuotation(lines),
		IdempotencyKey:    params.IdempotencyKey,
		OrgId:             stringOf(quotation, basemodel.FieldOrgId),
	}, taxSvc, products, basisSvc, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	if err := stampQuotationAccepted(ctx, quotation, created.SalesOrderId); err != nil {
		return nil, nil, err
	}

	order, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, created.SalesOrderId)
	if err != nil {
		return nil, nil, err
	}

	return &ConvertQuotationResult{
		SalesQuotationId: params.SalesQuotationId,
		SalesOrderId:     created.SalesOrderId,
		OrderNumber:      created.OrderNumber,
		QuotedTotal:      decimalOf(quotation, models.SalesQuotationFieldGrandTotal),
		OrderTotal:       decimalOf(order, models.SalesOrderFieldGrandTotal),
	}, nil, nil
}

// assertConvertible applies the two gates that stand between an offer and a sale.
func assertConvertible(quotation dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	status := stringOf(quotation, models.SalesQuotationFieldStatus)

	// Already-accepted is refused here rather than left to canTransition, which treats from == to as
	// an allowed no-op — right for an idempotent status retry, wrong when accepting creates an order.
	// This path is only reached when converted_sales_order_id is missing but the status is not, a
	// half-written accept, and refusing is safer than making a second sale.
	if status == string(models.SalesQuotationStatusAccepted) {
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonQuotationNotConvertible,
			"this quotation has already been accepted"))
		return vErrs
	}

	if !CanTransitionQuotation(status, string(models.SalesQuotationStatusAccepted)) {
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonQuotationNotConvertible,
			"a quotation in status '"+status+"' cannot be accepted"))
		return vErrs
	}

	// Expiry is checked here as well as by the sweep: the sweep runs on a schedule, so between a
	// quotation lapsing and the sweep noticing, the stored status still says sent, and converting in
	// that window would honour an expired offer.
	if validUntil := dateTimeOf(quotation, models.SalesQuotationFieldValidUntil); validUntil != nil {
		if validUntil.GoTime().Before(time.Now().UTC()) {
			vErrs.Append(*ft.NewBusinessViolation("valid_until", ReasonQuotationExpired,
				"this quotation lapsed on "+validUntil.GoTime().Format(time.RFC3339)+
					" and must be re-quoted rather than converted"))
		}
	}

	if vErrs.Count() > 0 {
		return vErrs
	}
	return nil
}

// orderLinesFromQuotation carries what was asked for, not what it was quoted at. The unit price
// travels as the catalogue fallback, which the engine overrides whenever a pricelist matches, so a
// quoted price survives only where nothing else has an opinion.
func orderLinesFromQuotation(lines []dmodel.DynamicFields) []CreateOrderLine {
	converted := make([]CreateOrderLine, 0, len(lines))
	for _, line := range lines {
		converted = append(converted, CreateOrderLine{
			ProductVariantId: stringOf(line, models.SalesQuotationLineFieldVariantId),
			UomId:            stringOf(line, models.SalesQuotationLineFieldUomId),
			Quantity:         decimalOf(line, models.SalesQuotationLineFieldQuantity),
			UnitPrice:        decimalOf(line, models.SalesQuotationLineFieldUnitPrice),
		})
	}
	return converted
}

// stampQuotationAccepted records the acceptance and the order it produced, both in one transaction:
// a quotation marked accepted with no order recorded is spent with nothing to show for it, and the
// reverse would let a second accept create a second order.
func stampQuotationAccepted(
	ctx corectx.Context, quotation dmodel.DynamicFields, orderId string,
) error {
	quotationId := stringOf(quotation, models.SalesQuotationFieldId)

	return withTransaction(ctx, models.SalesQuotationSchemaName,
		func(tranxCtx corectx.Context) error {
			engine, err := engineFor(models.SalesQuotationSchemaName)
			if err != nil {
				return err
			}
			if _, err := engine.ResourceRepository().Update(tranxCtx, dmodel.DynamicFields{
				models.SalesQuotationFieldId:             quotationId,
				models.SalesQuotationFieldStatus:         string(models.SalesQuotationStatusAccepted),
				models.SalesQuotationFieldConvertedOrder: orderId,
				models.SalesQuotationFieldAcceptedAt:     model.ModelDateTime(time.Now().UTC()),
			}); err != nil {
				return err
			}

			// The audit trail hangs off the order, the document that survives and the one an operator
			// investigating a price is looking at, and names the quotation it came from.
			return WriteSalesAuditEvent(tranxCtx, SalesAuditEntry{
				SalesOrderId: orderId,
				EntityType:   models.SalesQuotationSchemaName,
				EntityId:     quotationId,
				Action:       models.SalesOrderActionConvertQuotation,
				FromStatus:   stringOf(quotation, models.SalesQuotationFieldStatus),
				ToStatus:     string(models.SalesQuotationStatusAccepted),
				Reason: "converted from quotation " +
					stringOf(quotation, models.SalesQuotationFieldNumber),
				OrgId: stringOf(quotation, basemodel.FieldOrgId),
			})
		})
}

// replayConversion returns the order a previous accept already produced.
func replayConversion(
	ctx corectx.Context, quotation dmodel.DynamicFields,
) (*ConvertQuotationResult, *ft.ClientErrors, error) {
	orderId := stringOf(quotation, models.SalesQuotationFieldConvertedOrder)

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}

	result := &ConvertQuotationResult{
		SalesQuotationId: stringOf(quotation, models.SalesQuotationFieldId),
		SalesOrderId:     orderId,
		AlreadyConverted: true,
		QuotedTotal:      decimalOf(quotation, models.SalesQuotationFieldGrandTotal),
	}
	if order != nil {
		result.OrderNumber = stringOf(order, models.SalesOrderFieldOrderNumber)
		result.OrderTotal = decimalOf(order, models.SalesOrderFieldGrandTotal)
	}
	return result, nil, nil
}

// TransitionQuotation moves a quotation between statuses, refusing what the table forbids. The status
// field is declared no_update, so this is the only way a status moves at all.
func TransitionQuotation(
	ctx corectx.Context, quotationId, toStatus string,
) (*ft.ClientErrors, error) {
	quotation, err := loadRecord(ctx,
		models.SalesQuotationSchemaName, models.SalesQuotationFieldId, quotationId)
	if err != nil {
		return nil, err
	}
	if quotation == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_quotation_id", ReasonQuotationNotFound,
			"no quotation exists with id '"+quotationId+"'"))
		return vErrs, nil
	}

	fromStatus := stringOf(quotation, models.SalesQuotationFieldStatus)
	if !CanTransitionQuotation(fromStatus, toStatus) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonQuotationBadTransition,
			"a quotation cannot move from '"+fromStatus+"' to '"+toStatus+"'"))
		return vErrs, nil
	}

	// Accepting is ConvertQuotation's job: it must create an order and set converted_sales_order_id in
	// the same breath, and a status-only accept leaves a spent quotation with nothing to show for it.
	if toStatus == string(models.SalesQuotationStatusAccepted) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonQuotationBadTransition,
			"accept a quotation by converting it, so that the order it produces is recorded"))
		return vErrs, nil
	}

	engine, err := engineFor(models.SalesQuotationSchemaName)
	if err != nil {
		return nil, err
	}

	update := dmodel.DynamicFields{
		models.SalesQuotationFieldId:     quotationId,
		models.SalesQuotationFieldStatus: toStatus,
	}
	switch models.SalesQuotationStatus(toStatus) {
	case models.SalesQuotationStatusSent:
		update[models.SalesQuotationFieldSentAt] = model.ModelDateTime(time.Now().UTC())
	case models.SalesQuotationStatusCancelled:
		update[models.SalesQuotationFieldCancelledAt] = model.ModelDateTime(time.Now().UTC())
	}

	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return nil, err
	}
	return nil, nil
}
