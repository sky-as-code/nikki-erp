package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Repricing a draft purchase order (section 30).
//
// The operation exists BECAUSE a vendor price change must not rewrite an existing order by itself.
// Section 30 states that rule directly, and this action is its other half: the prices move when
// somebody asks for them to move, and the asking is recorded. An order that repriced itself when a
// master row changed would be a document whose numbers nobody chose.
//
// It is therefore deliberately explicit, deliberately audited, and deliberately refused on a
// committed order — a confirmed purchase order is a document the vendor holds a copy of, and
// changing its prices afterwards would make the two disagree with nothing to show why.

// RepricedLine is what happened to one line, for the caller's report.
type RepricedLine struct {
	LineId   string
	OldPrice decimal.Decimal
	NewPrice decimal.Decimal
}

// RepriceResultData reports what the operation did.
//
// The count of lines EXAMINED is reported beside the ones that changed, because "nothing changed"
// and "nothing was eligible" are different answers and a caller shown only a zero cannot tell them
// apart. One means the prices are already right; the other means the order has no product lines,
// or no vendor, and the operation did nothing at all.
type RepriceResultData struct {
	LinesExamined int
	Changed       []RepricedLine
}

// Reprice re-resolves every eligible line's price from the vendor's current price list (section 30).
//
// Eligible means a product line with a product on it. A section, a note or a free-text charge is
// skipped: there is no product to look up a vendor price for, and pricing one at zero would delete
// a freight charge somebody typed.
//
// A line whose price does not change is left completely alone — not rewritten with the same value,
// and not audited. Writing an unchanged row would touch its etag and its audit trail for no reason,
// and a trail full of no-op events is one nobody reads.
func (this *PurchaseOrderDomainServiceImpl) Reprice(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[RepriceResultData], error) {
	var result *dyn.OpResult[RepriceResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			result = &dyn.OpResult[RepriceResultData]{
				ClientErrors: *orderNotFoundErrors(orderId),
			}
			return nil
		}

		status := stringOf(order, models.PurchaseOrderFieldStatus)
		if refusal := repriceRefusal(order, status); refusal != nil {
			result = &dyn.OpResult[RepriceResultData]{ClientErrors: *refusal}
			return nil
		}

		report, err := this.repriceLines(tranxCtx, order)
		if err != nil {
			return err
		}

		if len(report.Changed) > 0 {
			if err := RecomputeOrderTotals(tranxCtx, orderId); err != nil {
				return err
			}
		}

		result = &dyn.OpResult[RepriceResultData]{HasData: true, Data: report}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// repriceRefusal reports why an order may not be repriced, or nil when it may.
//
// Two refusals, and they are different rules rather than two phrasings of one. A COMMITTED order is
// a document the vendor is holding: its prices are what was agreed, and moving them afterwards
// would make the two copies disagree. A LOCKED order is one somebody deliberately closed to
// editing, and repricing is an edit — the fact that it edits several lines at once rather than one
// is not an exemption.
func repriceRefusal(order dmodel.DynamicFields, status string) *ft.ClientErrors {
	if IsOrderCommitted(status) {
		return orderViolationErrors("purchase_order.not_repriceable",
			formatStatus("only a draft purchase order can be repriced; this one is '%s'", status))
	}
	if status == string(models.PurchaseOrderStatusCancelled) {
		return orderViolationErrors("purchase_order.cancelled_is_final",
			"a cancelled purchase order cannot be repriced; duplicate it to start a new request for quotation")
	}
	if boolOf(order, models.PurchaseOrderFieldIsLocked) {
		return orderViolationErrors("purchase_order.locked",
			"this purchase order is locked; unlock it before repricing")
	}
	return nil
}

// repriceLines re-resolves each eligible line and writes the ones that moved.
//
// Every line is priced against the SAME order record, read once above. Re-reading the vendor per
// line would let a concurrent change to the header split one repricing across two vendors, which is
// a document nobody could explain.
func (this *PurchaseOrderDomainServiceImpl) repriceLines(
	ctx corectx.Context, order dmodel.DynamicFields,
) (RepriceResultData, error) {
	report := RepriceResultData{}
	if this.pricer == nil || this.products == nil {
		// No ports means no way to resolve anything. Reported as "examined nothing" rather than as
		// an error: this happens only in a test that built the service by hand.
		return report, nil
	}

	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return report, err
	}
	orderId := stringOf(order, models.PurchaseOrderFieldId)
	lines, err := models.FindOrderLines(ctx, lineEngine.ResourceRepository(),
		orderId, models.MaxOrderLines)
	if err != nil {
		return report, err
	}

	priceAt := timeNow()
	for _, line := range lines {
		if !isMoneyBearingLine(line) {
			continue
		}
		if stringOf(line, models.PurchaseOrderLineFieldProductVariantId) == "" {
			// A free-text charge — freight, a one-off fee. There is no product whose vendor price
			// could be looked up, and zeroing it would delete a real cost from the order.
			continue
		}
		report.LinesExamined++

		changed, err := this.repriceOneLine(ctx, line, order, priceAt)
		if err != nil {
			return report, err
		}
		if changed != nil {
			report.Changed = append(report.Changed, *changed)
		}
	}
	return report, nil
}

// repriceOneLine re-resolves a single line, writing and auditing it only if the price moved.
//
// The product is re-validated on the way through — that is what supplies the template id, and it is
// also the right moment to notice that a product has since been archived or made unpurchasable. A
// line for such a product is SKIPPED rather than refused: the order already contains it, refusing
// the whole operation would strand every other line, and the buyer can see the problem on the line
// itself.
func (this *PurchaseOrderDomainServiceImpl) repriceOneLine(
	ctx corectx.Context, line, order dmodel.DynamicFields, priceAt time.Time,
) (*RepricedLine, error) {
	probe := dmodel.DynamicFields{}
	for key, value := range line {
		probe[key] = value
	}

	vErrs := ft.NewClientErrors()
	templateId, err := this.products.PrepareLine(ctx, probe, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 || templateId == "" {
		return nil, nil
	}

	// Deleted so the pricer treats the price as unstated and fills in the vendor's current one.
	// Without this the line's existing price would be read as a negotiated one and left alone,
	// which is exactly right on an ordinary save and exactly wrong here: repricing is the operation
	// that deliberately asks for the current number.
	delete(probe, models.PurchaseOrderLineFieldUnitPrice)
	if err := this.pricer.PriceLine(ctx, probe, order, templateId, priceAt); err != nil {
		return nil, err
	}

	if stringOf(probe, models.PurchaseOrderLineFieldVendorProductPriceId) == "" {
		// Nothing is quoted for this product any more. The line keeps the price it has — section 28
		// forbids substituting a cost, and zeroing it would give the goods away. A withdrawn quote
		// is not a reason to change what an order says it will pay.
		return nil, nil
	}

	oldPrice := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)
	newPrice := decimalOf(probe, models.PurchaseOrderLineFieldUnitPrice)
	priceMoved := !oldPrice.Equal(newPrice)
	referenceMoved := stringOf(line, models.PurchaseOrderLineFieldVendorProductPriceId) !=
		stringOf(probe, models.PurchaseOrderLineFieldVendorProductPriceId)
	if !priceMoved && !referenceMoved {
		return nil, nil
	}

	StampLineTotals(probe)
	if err := this.writeRepricedLine(ctx, line, probe); err != nil {
		return nil, err
	}

	lineId := stringOf(line, models.PurchaseOrderLineFieldId)
	if err := WriteAuditEvent(ctx, AuditEntry{
		EntityType: models.PurchaseOrderLineSchemaName,
		EntityId:   lineId,
		Action:     AuditActionReprice,
		OrgId:      stringOf(line, basemodel.FieldOrgId),
		Metadata: map[string]any{
			"purchase_order_id":       stringOf(line, models.PurchaseOrderLineFieldPurchaseOrderId),
			"old_unit_price":          oldPrice.String(),
			"new_unit_price":          newPrice.String(),
			"vendor_product_price_id": stringOf(probe, models.PurchaseOrderLineFieldVendorProductPriceId),
		},
	}); err != nil {
		return nil, err
	}

	if !priceMoved {
		// The reference moved but the number did not: two quotes at the same price, one of which
		// has since expired. Worth auditing, since the line now points somewhere else, but not
		// worth reporting as a price change to a caller who would see old and new be equal.
		return nil, nil
	}
	return &RepricedLine{LineId: lineId, OldPrice: oldPrice, NewPrice: newPrice}, nil
}

// writeRepricedLine saves the five fields repricing may move.
//
// Named explicitly rather than writing the whole record back, because most of a line is none of
// this operation's business: the quantity, the unit, the description and the expected arrival are
// the buyer's, and a blanket write would resubmit them all and could resurrect a value that another
// request changed between the read and the write.
func (this *PurchaseOrderDomainServiceImpl) writeRepricedLine(
	ctx corectx.Context, stored, priced dmodel.DynamicFields,
) error {
	engine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldId:                   stringOf(stored, models.PurchaseOrderLineFieldId),
		basemodel.FieldEtag:                               stringOf(stored, basemodel.FieldEtag),
		models.PurchaseOrderLineFieldUnitPrice:            priced[models.PurchaseOrderLineFieldUnitPrice],
		models.PurchaseOrderLineFieldVendorProductPriceId: priced[models.PurchaseOrderLineFieldVendorProductPriceId],
		models.PurchaseOrderLineFieldResolvedUnitPrice:    priced[models.PurchaseOrderLineFieldResolvedUnitPrice],
		models.PurchaseOrderLineFieldSubtotal:             priced[models.PurchaseOrderLineFieldSubtotal],
		models.PurchaseOrderLineFieldTotal:                priced[models.PurchaseOrderLineFieldTotal],
	}

	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeRepricedLine")
}
