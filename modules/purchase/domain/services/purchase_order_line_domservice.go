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

// The line service keeps the stored header totals equal to the sum of the lines: every create,
// update and delete recomputes the header before returning. It is a derived service rather than an
// engine callback because the recompute writes to another resource and must share the write's
// transaction, and the ValidateExtra hook only runs before the write.

// NewPurchaseOrderLineDomainService wraps base, the line engine's own resource service, so built-in
// CRUD still runs through it and the totals work layers around it. Install with
// Engine.SetResourceService.
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
	// Nil in tests that exercise only the totals arithmetic, which needs no ports.
	products *ProductLineValidator

	// pricer resolves the vendor price for a line. Nil in the same tests; a nil pricer leaves the
	// line with whatever price it was given, as happens in production when no quote applies.
	pricer *LinePricer
}

var _ drif.DynamicResourceService = (*PurchaseOrderLineDomainServiceImpl)(nil)

// Create stamps the line's computed money fields, then recomputes the header. The stamp must
// precede the base call: subtotal and total are required_for_create, so the schema would refuse the
// create before this service could repair it.
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

		// Pricing runs after validation but before the write, so the resolved price and its
		// reference are part of the same insert.
		if err := this.priceLine(tranxCtx, params, templateId); err != nil {
			return err
		}
		// Restamped because pricing may have supplied the unit price the first stamp lacked.
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

// Update recomputes the line's money fields and then the header's. The stored line is read and
// merged under the incoming params first because an update is partial: computing a subtotal from
// the params alone would price a quantity-only change at zero.
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

			// An update never re-prices — that would undo a negotiation on every save. It only
			// records that the sent price differs from the resolved one.
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

// Delete removes the line and recomputes the header. The owning order must be read before the
// delete: afterwards the line, and with it the only pointer to the order, is gone.
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

// prepareProduct applies the product and unit rules, filling inventory_quantity. A nil validator
// means the ports were never bound (unit tests of the arithmetic alone); the field is then stamped
// from the ordered quantity so required_for_create is still satisfied.
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

// mergeStoredLine returns the stored line with the incoming params laid over it, plus the owning
// order id. A missing line is not an error here; the base call that follows reports it.
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

// priceLine resolves the vendor price for a line about to be created. It reads the owning order
// only for the vendor id, which lives on the header. A nil pricer, an unreadable order or an order
// with no vendor all leave the line as it arrived: pricing is fail-soft, not a gate.
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

// auditPriceOverride records that a line was priced differently from the vendor's quote. It
// re-resolves into a throwaway copy, never the line itself, so the update notices the difference
// without silently rewriting the buyer's price. The quote reference and resolved price are still
// refreshed on the line, or a stale reference would misattribute the override to a dead quote.
func (this *PurchaseOrderLineDomainServiceImpl) auditPriceOverride(
	ctx corectx.Context, merged, params dmodel.DynamicFields, templateId string,
) error {
	if this.pricer == nil || templateId == "" {
		return nil
	}
	// Only a request that carries a price can be an override; auditing a quantity-only update
	// would fill the trail with events nobody caused.
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
		// Nothing was quoted, so there is nothing to have overridden.
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

// readOwningOrder fetches the header a line belongs to, returning nil rather than an error for a
// missing order: the base call that follows already refuses a line naming an order that is not there.
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
