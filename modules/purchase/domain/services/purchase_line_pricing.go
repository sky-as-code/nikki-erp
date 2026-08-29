package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services/vendorpricing"
)

// Purchase price resolution, the impure half: vendorpricing ranks the candidates, this file reads
// them, dates the windows, converts quantities and stamps the line, so the ranking rules stay
// testable without a database. It must never invent a price — when no vendor price applies the line
// keeps the typed price and records no reference, with no fallback to product cost or another
// vendor's offer.

// LinePricer resolves the vendor price for a purchase order line. It holds the unit port because
// unit conversion is Essential's job, never this module's; it holds no product port because the
// line service has already resolved the product before pricing runs.
type LinePricer struct {
	uoms itExt.UomExtService
}

func NewLinePricer(uoms itExt.UomExtService) *LinePricer {
	return &LinePricer{uoms: uoms}
}

// PriceLine stamps vendor_product_price_id (which quote) and resolved_unit_price (what it said),
// but overwrites unit_price only when the caller stated none — an explicit price is a negotiated
// one, and the resolved figure is recorded beside it so the override stays auditable. Failing to
// resolve a price is never a failure to save the line.
func (this *LinePricer) PriceLine(
	ctx corectx.Context, line dmodel.DynamicFields, order dmodel.DynamicFields,
	templateId string, priceAt time.Time,
) error {
	if !isMoneyBearingLine(line) {
		return nil
	}

	vendorId := stringOf(order, models.PurchaseOrderFieldVendorId)
	orgId := stringOf(line, models.PurchaseOrderLineFieldOrgId)
	if orgId == "" {
		orgId = stringOf(order, models.PurchaseOrderFieldOrgId)
	}
	if vendorId == "" || orgId == "" || templateId == "" {
		// A draft without a vendor, or a free-text line without a product, cannot be priced from a
		// vendor price list; neither is a fault.
		return nil
	}

	candidates, err := this.loadCandidates(ctx, orgId, vendorId, templateId, priceAt)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	request, err := this.buildRequest(ctx, line, templateId, candidates)
	if err != nil {
		return err
	}

	resolved, found := vendorpricing.Resolve(request, candidates)
	if !found {
		// Not an error and not a fallback: the line keeps its typed price and records no reference.
		return nil
	}

	line[models.PurchaseOrderLineFieldVendorProductPriceId] = resolved.VendorProductPriceId
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = resolved.UnitPrice
	if !hasExplicitPrice(line) {
		line[models.PurchaseOrderLineFieldUnitPrice] = resolved.UnitPrice
	}
	return nil
}

// loadCandidates reads this vendor's quotes for this product and stamps each window verdict against
// priceAt. Every candidate must be judged against the same instant; reading the clock per row would
// let a line saved across midnight rank two quotes by two different dates.
func (this *LinePricer) loadCandidates(
	ctx corectx.Context, orgId, vendorId, templateId string, priceAt time.Time,
) ([]vendorpricing.Candidate, error) {
	engine, err := engineFor(models.VendorProductPriceSchemaName)
	if err != nil {
		return nil, err
	}
	rows, err := models.FindVendorPriceCandidates(ctx, engine.ResourceRepository(),
		orgId, vendorId, templateId, models.MaxVendorPriceCandidates)
	if err != nil {
		return nil, errors.Wrap(err, "loadCandidates")
	}

	candidates := make([]vendorpricing.Candidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, vendorpricing.Candidate{
			Id:                stringOf(row, models.VendorProductPriceFieldId),
			ProductTemplateId: stringOf(row, models.VendorProductPriceFieldProductTemplateId),
			ProductVariantId:  stringOf(row, models.VendorProductPriceFieldProductVariantId),
			PurchaseUomId:     stringOf(row, models.VendorProductPriceFieldPurchaseUomId),
			CurrencyId:        stringOf(row, models.VendorProductPriceFieldCurrencyId),
			MinQuantity:       decimalOf(row, models.VendorProductPriceFieldMinQuantity),
			UnitPrice:         decimalOf(row, models.VendorProductPriceFieldUnitPrice),
			Applicable:        windowCovers(row, priceAt),
			LeadTimeDays:      int32Of(row, models.VendorProductPriceFieldLeadTimeDays),
			Sequence:          int32Of(row, models.VendorProductPriceFieldSequence),
		})
	}
	return candidates, nil
}

// buildRequest expresses the ordered quantity once per distinct quoting unit, not once per
// candidate. A unit that fails to convert is left out of the map so vendorpricing skips it;
// storing a zero instead would make every quantity break in that unit look reachable.
func (this *LinePricer) buildRequest(
	ctx corectx.Context, line dmodel.DynamicFields, templateId string,
	candidates []vendorpricing.Candidate,
) (vendorpricing.Request, error) {
	quantity := decimalOf(line, models.PurchaseOrderLineFieldQuantity)
	lineUomId := stringOf(line, models.PurchaseOrderLineFieldUomId)

	quantities := make(map[string]decimal.Decimal, len(candidates))
	for _, candidate := range candidates {
		unit := candidate.PurchaseUomId
		if unit == "" {
			continue
		}
		if _, done := quantities[unit]; done {
			continue
		}
		converted, ok, err := this.convert(ctx, quantity, lineUomId, unit)
		if err != nil {
			return vendorpricing.Request{}, err
		}
		if ok {
			quantities[unit] = converted
		}
	}

	return vendorpricing.Request{
		ProductTemplateId: templateId,
		ProductVariantId:  stringOf(line, models.PurchaseOrderLineFieldProductVariantId),
		QuantityByUom:     quantities,
	}, nil
}

// convert asks Essential for the quantity in another unit, reporting a refusal as "not available"
// rather than as an error: a cross-category refusal (a per-kilogram quote against a line in litres)
// is a correct answer, and erroring would refuse to save the line over an unused quote.
func (this *LinePricer) convert(
	ctx corectx.Context, quantity decimal.Decimal, sourceUomId, targetUomId string,
) (decimal.Decimal, bool, error) {
	if targetUomId == sourceUomId || sourceUomId == "" {
		// No unit on the line: treat the bare quantity as already in the quote's unit.
		return quantity, true, nil
	}
	if this.uoms == nil {
		return decimal.Zero, false, nil
	}

	converted, err := this.uoms.Convert(ctx, itExt.ConvertQuantityQuery{
		Quantity:    quantity,
		SourceUomId: model.Id(sourceUomId),
		TargetUomId: model.Id(targetUomId),
	})
	if err != nil {
		return decimal.Zero, false, errors.Wrap(err, "convert")
	}
	if converted == nil || !converted.HasData || converted.ClientErrors.Count() > 0 {
		return decimal.Zero, false, nil
	}
	return converted.Data.Quantity, true, nil
}

// windowCovers reports whether a quote's window is open at the given instant. Both bounds are
// optional and inclusive; an absent bound is open-ended. It must read through timeOf, not
// DynamicFields.GetModelDateTime, which returns nil for shapes it does not recognise (including the
// *ModelDateTime the model layer's own setter produces) — nil reads as "no bound" and would
// silently turn an expired quote into a standing offer.
func windowCovers(row dmodel.DynamicFields, at time.Time) bool {
	if from, ok := timeOf(row, models.VendorProductPriceFieldValidFrom); ok && at.Before(from) {
		return false
	}
	if to, ok := timeOf(row, models.VendorProductPriceFieldValidTo); ok && at.After(to) {
		return false
	}
	return true
}

// hasExplicitPrice tests key presence, not the value: zero is a legitimate price for a
// free-of-charge line, so treating it as absent would silently overwrite one.
func hasExplicitPrice(line dmodel.DynamicFields) bool {
	value, present := line[models.PurchaseOrderLineFieldUnitPrice]
	return present && value != nil
}

// int32Of reads an int32 field, tolerating the several shapes a repository may hand back.
func int32Of(fields dmodel.DynamicFields, key string) int32 {
	value, ok := fields[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return typed
	case *int32:
		if typed == nil {
			return 0
		}
		return *typed
	case int:
		return int32(typed)
	case int64:
		return int32(typed)
	case float64:
		return int32(typed)
	default:
		return 0
	}
}
