package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The line service is where "the stored totals are always the totals of the lines" is actually
// made true. Every write to a line — create, update or delete — changes what the header should
// say, so every one of them recomputes it before returning.
//
// It is a derived service rather than an engine callback because the recompute writes to another
// resource, and because it must share the write's transaction: the ValidateExtra hook runs before
// the write and has nowhere to put an after.

// NewPurchaseOrderLineDomainService derives the line service from the engine's default one.
//
// base is the line engine's own resource service, which this type embeds: built-in CRUD keeps
// running through the default implementation, and the totals work is layered around it. The result
// is installed with Engine.SetResourceService.
func NewPurchaseOrderLineDomainService(
	base drif.DynamicResourceService, products *ProductLineValidator, pricer *LinePricer,
) *PurchaseOrderLineDomainServiceImpl {
	return &PurchaseOrderLineDomainServiceImpl{
		DynamicResourceService: base,
		products:               products,
		pricer:                 pricer,
	}
}

type PurchaseOrderLineDomainServiceImpl struct {
	drif.DynamicResourceService

	// products applies the product and unit-of-measure rules and computes inventory_quantity.
	// It is nil in tests that exercise only the totals arithmetic, which needs no ports.
	products *ProductLineValidator

	// pricer resolves the vendor price for a line (sections 27 and 29). Nil in the same tests, and
	// a nil pricer simply means the line keeps whatever price it was given — which is exactly what
	// happens in production too when no vendor quote applies.
	pricer *LinePricer
}

var _ drif.DynamicResourceService = (*PurchaseOrderLineDomainServiceImpl)(nil)

// Create stamps the line's computed money fields, then brings the header back in step.
//
// The stamp has to happen before the base call: subtotal and total are required_for_create, so a
// create without them is refused by the schema before this service could repair it.
func (this *PurchaseOrderLineDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	StampLineTotals(params)

	var result *dyn.OpResult[dmodel.DynamicFields]
	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		vErrs := ft.NewClientErrors()
		templateId, err := this.prepareProduct(tranxCtx, params, vErrs)
		if err != nil {
			return err
		}
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
			return nil
		}

		// Pricing runs BEFORE the write and after validation, so the resolved price and its
		// reference are part of the same insert. Doing it afterwards would leave a window in which
		// the line existed with a price nobody had resolved.
		if err := this.priceLine(tranxCtx, params, templateId); err != nil {
			return err
		}
		// Restamped because pricing may have supplied the unit price, and the totals stamped at the
		// top of this method were computed before it existed.
		StampLineTotals(params)

		created, err := this.DynamicResourceService.Create(tranxCtx, params)
		if err != nil || created.ClientErrors.Count() > 0 {
			result = created
			return err
		}
		result = created
		return RecomputeOrderTotals(tranxCtx, stringOf(params, models.PurchaseOrderLineFieldPurchaseOrderId))
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update recomputes the line's own money fields and then the header's.
//
// The line is read first because an update is partial: a request that changes only the quantity
// carries no unit price, and computing a subtotal from the params alone would price it at zero.
// The read merges the stored line under the incoming changes so the computation sees the whole
// line as it will be.
func (this *PurchaseOrderLineDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		merged, orderId, err := this.mergeStoredLine(tranxCtx, params)
		if err != nil {
			return err
		}
		if merged != nil {
			vErrs := ft.NewClientErrors()
			templateId, err := this.prepareProduct(tranxCtx, merged, vErrs)
			if err != nil {
				return err
			}
			if vErrs.Count() > 0 {
				result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
				return nil
			}

			// An update never re-prices: whatever the buyer sent is what the line says. What it
			// does do is notice that the sent price differs from what was resolved, and record
			// that (section 29.1). Re-pricing here would undo a negotiation on every save.
			if err := this.auditPriceOverride(tranxCtx, merged, params, templateId); err != nil {
				return err
			}

			StampLineTotals(merged)
			params[models.PurchaseOrderLineFieldSubtotal] = merged[models.PurchaseOrderLineFieldSubtotal]
			params[models.PurchaseOrderLineFieldTaxAmount] = merged[models.PurchaseOrderLineFieldTaxAmount]
			params[models.PurchaseOrderLineFieldTotal] = merged[models.PurchaseOrderLineFieldTotal]
			// inventory_quantity is no_update, so it is written through the repository by the
			// recompute rather than carried in the client's update params.
			params[models.PurchaseOrderLineFieldInventoryQuantity] =
				merged[models.PurchaseOrderLineFieldInventoryQuantity]
		}

		updated, err := this.DynamicResourceService.Update(tranxCtx, params)
		result = updated
		if err != nil || updated.ClientErrors.Count() > 0 {
			return err
		}
		if orderId == "" {
			return nil
		}
		return RecomputeOrderTotals(tranxCtx, orderId)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Delete removes the line and then brings the header back in step.
//
// The owning order is read BEFORE the delete, because afterwards the line is gone and with it the
// only pointer to the order whose totals just changed.
func (this *PurchaseOrderLineDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		_, orderId, err := this.mergeStoredLine(tranxCtx, params)
		if err != nil {
			return err
		}

		deleted, err := this.DynamicResourceService.Delete(tranxCtx, params)
		result = deleted
		if err != nil || deleted.ClientErrors.Count() > 0 {
			return err
		}
		if orderId == "" {
			return nil
		}
		return RecomputeOrderTotals(tranxCtx, orderId)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// prepareProduct applies the product and unit rules, filling inventory_quantity.
//
// A nil validator means the ports were never bound, which happens only in a unit test that
// exercises the arithmetic alone. In that case the field is stamped from the ordered quantity so
// the required_for_create constraint is still satisfied.
func (this *PurchaseOrderLineDomainServiceImpl) prepareProduct(
	ctx corectx.Context, line dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (string, error) {
	if this.products == nil {
		if _, present := line[models.PurchaseOrderLineFieldInventoryQuantity]; !present {
			line[models.PurchaseOrderLineFieldInventoryQuantity] =
				decimalOf(line, models.PurchaseOrderLineFieldQuantity)
		}
		return "", nil
	}
	return this.products.PrepareLine(ctx, line, vErrs)
}

// mergeStoredLine reads the stored line and returns it with the incoming params laid over the top,
// along with the id of the order that owns it.
//
// A missing line is not an error here: the base call that follows reports it, and reporting it
// twice in two different shapes would only make the response harder to read.
func (this *PurchaseOrderLineDomainServiceImpl) mergeStoredLine(
	ctx corectx.Context, params dmodel.DynamicFields,
) (dmodel.DynamicFields, string, error) {
	lineId := stringOf(params, models.PurchaseOrderLineFieldId)
	if lineId == "" {
		return nil, "", nil
	}

	engine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return nil, "", err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PurchaseOrderLineFieldId: lineId,
	})
	if err != nil {
		return nil, "", errors.Wrap(err, "mergeStoredLine")
	}
	if found == nil || !found.HasData {
		return nil, "", nil
	}

	merged := make(dmodel.DynamicFields, len(found.Data)+len(params))
	for key, value := range found.Data {
		merged[key] = value
	}
	for key, value := range params {
		merged[key] = value
	}
	// purchase_order_id is no_update, so the stored value is the authority even if the request
	// carried a different one — following the params here would recompute the wrong order.
	return merged, stringOf(found.Data, models.PurchaseOrderLineFieldPurchaseOrderId), nil
}

// priceLine resolves the vendor price for a line about to be created.
//
// The owning order is read for one reason only: the vendor. A price is a vendor's offer, so
// resolution cannot begin without knowing whose offer to look for, and the vendor lives on the
// header rather than on the line.
//
// A nil pricer, an unreadable order or an order with no vendor all leave the line exactly as it
// arrived. That is the same fail-soft stance PrepareLine takes for a product with no stock
// configuration: pricing is a service to the buyer, not a gate they must pass.
func (this *PurchaseOrderLineDomainServiceImpl) priceLine(
	ctx corectx.Context, line dmodel.DynamicFields, templateId string,
) error {
	if this.pricer == nil || templateId == "" {
		return nil
	}

	order, err := this.readOwningOrder(ctx, line)
	if err != nil || order == nil {
		return err
	}
	return this.pricer.PriceLine(ctx, line, order, templateId, timeNow())
}

// auditPriceOverride records that somebody priced a line differently from the vendor's quote.
//
// It re-resolves into a THROWAWAY copy of the line rather than into the line itself, which is the
// whole point: the update must not change the price, only notice that it differs. Writing the
// resolved figure back would be the silent rewrite section 30 forbids.
//
// The reference and the resolved price are refreshed on the stored line even so, because they
// describe which quote the comparison was made against — leaving a stale reference beside a new
// price would misattribute the override to a quote that no longer applies.
func (this *PurchaseOrderLineDomainServiceImpl) auditPriceOverride(
	ctx corectx.Context, merged, params dmodel.DynamicFields, templateId string,
) error {
	if this.pricer == nil || templateId == "" {
		return nil
	}
	// Only a request that actually carries a price can be an override. An update of the quantity
	// alone leaves the price untouched, and auditing one would fill the trail with events nobody
	// caused.
	if _, sent := params[models.PurchaseOrderLineFieldUnitPrice]; !sent {
		return nil
	}

	order, err := this.readOwningOrder(ctx, merged)
	if err != nil || order == nil {
		return err
	}

	probe := dmodel.DynamicFields{}
	for key, value := range merged {
		probe[key] = value
	}
	// Deleted so the pricer treats the price as unstated and fills in what the vendor quotes;
	// merged still holds what the buyer sent, and that is what will be saved.
	delete(probe, models.PurchaseOrderLineFieldUnitPrice)
	if err := this.pricer.PriceLine(ctx, probe, order, templateId, timeNow()); err != nil {
		return err
	}

	resolvedId := stringOf(probe, models.PurchaseOrderLineFieldVendorProductPriceId)
	merged[models.PurchaseOrderLineFieldVendorProductPriceId] = resolvedId
	params[models.PurchaseOrderLineFieldVendorProductPriceId] = resolvedId
	merged[models.PurchaseOrderLineFieldResolvedUnitPrice] =
		probe[models.PurchaseOrderLineFieldResolvedUnitPrice]
	params[models.PurchaseOrderLineFieldResolvedUnitPrice] =
		probe[models.PurchaseOrderLineFieldResolvedUnitPrice]

	if resolvedId == "" {
		// Nothing was quoted, so there is nothing to have overridden. A hand-typed price on a
		// product no vendor has quoted is an ordinary line, not an exception.
		return nil
	}
	resolved := decimalOf(probe, models.PurchaseOrderLineFieldUnitPrice)
	agreed := decimalOf(merged, models.PurchaseOrderLineFieldUnitPrice)
	if resolved.Equal(agreed) {
		return nil
	}

	return WriteAuditEvent(ctx, AuditEntry{
		EntityType: models.PurchaseOrderLineSchemaName,
		EntityId:   stringOf(merged, models.PurchaseOrderLineFieldId),
		Action:     AuditActionOverridePrice,
		OrgId:      stringOf(merged, models.PurchaseOrderLineFieldOrgId),
		Metadata: map[string]any{
			"vendor_product_price_id": resolvedId,
			"resolved_unit_price":     resolved.String(),
			"unit_price":              agreed.String(),
			"purchase_order_id": stringOf(
				merged, models.PurchaseOrderLineFieldPurchaseOrderId),
		},
	})
}

// readOwningOrder fetches the header a line belongs to.
//
// A missing order is reported as nil rather than as an error: on create the base call that follows
// will refuse the line for naming an order that does not exist, and producing a second, differently
// worded refusal here would only make the response harder to read.
func (this *PurchaseOrderLineDomainServiceImpl) readOwningOrder(
	ctx corectx.Context, line dmodel.DynamicFields,
) (dmodel.DynamicFields, error) {
	orderId := stringOf(line, models.PurchaseOrderLineFieldPurchaseOrderId)
	if orderId == "" {
		return nil, nil
	}

	engine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PurchaseOrderFieldId: orderId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "readOwningOrder")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}
