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

// Repricing a draft purchase order. A vendor price change must never rewrite an existing order by
// itself, so this operation is explicit and audited: prices move only when somebody asks. It is
// refused on a committed order, of which the vendor holds a copy that would otherwise disagree.

// RepricedLine is what happened to one line, for the caller's report.
type RepricedLine struct {
	LineId   string
	OldPrice decimal.Decimal
	NewPrice decimal.Decimal
}

// RepriceResultData reports what the operation did. It counts lines examined beside the ones that
// changed, so a caller can tell "nothing changed" from "nothing was eligible".
type RepriceResultData struct {
	LinesExamined int
	Changed       []RepricedLine
}

// Reprice re-resolves every eligible line's price from the vendor's current price list. Eligible
// means a product line with a product on it; sections, notes and free-text charges are skipped,
// since pricing one at zero would delete a charge somebody typed. A line whose price does not move
// is neither rewritten nor audited, so its etag and trail stay untouched.
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

// repriceRefusal reports why an order may not be repriced, or nil when it may. Committed and locked
// are separate rules: the first protects prices the vendor already holds a copy of, the second is
// the ordinary no-editing rule, from which repricing is not exempt.
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

// repriceLines re-resolves each eligible line and writes the ones that moved. Every line is priced
// against the same order record, read once by the caller: re-reading the vendor per line would let
// a concurrent header change split one repricing across two vendors.
func (this *PurchaseOrderDomainServiceImpl) repriceLines(
	ctx corectx.Context, order dmodel.DynamicFields,
) (RepriceResultData, error) {
	report := RepriceResultData{}
	if this.pricer == nil || this.products == nil {
		// No ports means nothing can be resolved; reported as "examined nothing" rather than an
		// error, which happens only in a test that built the service by hand.
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
			// A free-text charge has no product to look a vendor price up for, and zeroing it
			// would delete a real cost from the order.
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

// repriceOneLine re-resolves a single line, writing and auditing it only if the price moved. The
// product is re-validated on the way through, which supplies the template id and catches a product
// since archived or made unpurchasable; such a line is skipped rather than refused, so one bad line
// does not strand the rest of the order.
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
	// Left in place it would read as a negotiated price and be kept — right on an ordinary save,
	// wrong here, where the caller has asked for the current number.
	delete(probe, models.PurchaseOrderLineFieldUnitPrice)
	if err := this.pricer.PriceLine(ctx, probe, order, templateId, priceAt); err != nil {
		return nil, err
	}

	if stringOf(probe, models.PurchaseOrderLineFieldVendorProductPriceId) == "" {
		// Nothing is quoted for this product any more, so the line keeps its price: substituting a
		// cost is forbidden and zeroing it would give the goods away.
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
		// The reference moved but the number did not — audited, since the line now points
		// elsewhere, but not reported as a price change with equal old and new values.
		return nil, nil
	}
	return &RepricedLine{LineId: lineId, OldPrice: oldPrice, NewPrice: newPrice}, nil
}

// writeRepricedLine saves only the fields repricing may move. They are named explicitly rather than
// writing the whole record back, which would resubmit the buyer's quantity, unit and description and
// could resurrect a value another request changed between the read and the write.
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
