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

// Quotations and their conversion to orders (BR 87.1, SALES-038).
//
// # The conversion RE-PRICES rather than copying the quoted numbers
//
// This is the decision in this file, and the tempting alternative is wrong. Copying the quoted
// amounts onto the order would honour the offer exactly — which sounds right, and is how a quotation
// is often described — but it makes the order's numbers unexplainable: SALES-021's price explanation
// rebuilds a total from its adjustment chain, and a copied total has no chain. It would also let a
// quotation from six months ago create an order at prices, taxes and promotions that no longer
// exist, with nothing recording that it had happened.
//
// So the conversion runs the same engine a new order runs, and the quotation's role is to carry the
// LINES — what the customer asked for — not the money. `valid_until` is what makes that honest: an
// offer inside its window prices to substantially the same numbers, and one outside it is refused
// rather than silently repriced. Where the business genuinely wants to hold a stale price, that is a
// manual discount on the resulting order (SALES-039), which is permission-gated, reasoned and
// audited — all things a silent copy would not be.
//
// # Conversion is idempotent through converted_sales_order_id
//
// A second accept returns the first accept's order. A customer who accepted once agreed to buy once,
// and two orders from one acceptance is two deliveries and two invoices. The column is set inside
// the same transaction as the status move, so the two cannot disagree.

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
	// the quotation, because a quotation may be prepared centrally with no point decided — and
	// because the order's anti-spoofing rule (D-20) derives the channel from the POINT, so the point
	// is the thing that must be current.
	SalesPointId string

	// IdempotencyKey is passed through to the order. Belt to converted_sales_order_id's braces:
	// that column stops a second accept, and this stops a retry of the first one racing itself.
	IdempotencyKey string
}

// ConvertQuotationResult is what the conversion produced.
type ConvertQuotationResult struct {
	SalesQuotationId string
	SalesOrderId     string
	OrderNumber      string

	// AlreadyConverted marks the replay path: the quotation had already become an order, and that
	// order is what came back.
	AlreadyConverted bool

	// QuotedTotal and OrderTotal are BOTH returned, so a caller can see whether repricing moved the
	// number rather than taking it on trust. A back-office operator handing an order to a customer
	// who is holding the quotation needs to know before they do, not after.
	QuotedTotal decimal.Decimal
	OrderTotal  decimal.Decimal
}

// ConvertQuotation turns an accepted offer into a sales order (BR 87.1).
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

	// The replay check comes first, before the gates. A retry of a conversion that already happened
	// must succeed even if the quotation has since expired: the sale was made, and refusing the
	// acknowledgement would leave the caller retrying against an order that already exists.
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

	// The order is created through the ordinary path, not a private one. That is deliberate: it
	// keeps the anti-spoofing rule, the idempotency handling and the repricing identical to a
	// directly-created order, so a converted order cannot be a second class of order with its own
	// bugs.
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

	// Already-accepted is refused HERE, explicitly, rather than left to canTransition. That helper
	// treats from == to as an allowed no-op, which is right for an idempotent status retry and wrong
	// for this: accepting has a SIDE EFFECT, and a second one would create a second order.
	//
	// ConvertQuotation normally never reaches this point for an accepted quotation, because the
	// converted_sales_order_id replay catches it first. This is the case where that column is
	// missing but the status is not - a half-written accept - and the safe reading of that is to
	// refuse rather than to make a second sale.
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

	// Expiry is checked HERE as well as by the sweep, and the redundancy is the point: the sweep
	// runs on a schedule, so between a quotation lapsing and the sweep noticing there is a window in
	// which the stored status still says `sent`. Converting inside that window would honour an offer
	// that had already expired, at prices nobody re-agreed to.
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

// orderLinesFromQuotation carries WHAT was asked for, and not what it was quoted at.
//
// The unit price travels as the caller's catalogue fallback, which the engine overrides whenever a
// pricelist matches — so a quoted price survives only where nothing else has an opinion. That is the
// intended behaviour: see the package comment on why the conversion reprices.
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

// stampQuotationAccepted records the acceptance and the order it produced.
//
// Both writes in ONE transaction. A quotation marked accepted with no order recorded would be
// unconvertible and unexplainable — the offer is spent and nothing says what it bought — and the
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

			// The audit trail hangs off the ORDER, because that is the document that survives and
			// the one an operator investigating a price will be looking at. It names the quotation,
			// so the lineage reads in the direction somebody actually asks about it.
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

// TransitionQuotation moves a quotation between statuses, refusing what the table forbids.
//
// The one entry point for a quotation's status, so the transition table is enforced rather than
// merely documented — the field is declared no_update, so this is the only way a status moves at all.
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

	// Accepting is NOT done here. It has to create an order and set converted_sales_order_id in the
	// same breath, which is ConvertQuotation's job — and a status-only accept would leave a spent
	// quotation with nothing to show for it.
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
