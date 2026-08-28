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

// Purchase price resolution, the impure half (sections 27 and 29).
//
// vendorpricing decides; this file does the four things a decision cannot do for itself: read the
// candidates, ask the clock whether each window is open, convert the ordered quantity into every
// unit the candidates quote in, and write the answer onto the line. Splitting it this way is what
// lets the ranking rules be tested without a database, and it mirrors how Sales already separates
// its pricing engine from the service that feeds it.
//
// What this file must NOT do is invent a price. When no vendor price applies the line keeps
// whatever the buyer typed and records no reference — section 28 and TS-PRICE-10 are explicit that
// there is no fallback to the product cost or to another vendor offer.

// LinePricer resolves the vendor price for a purchase order line.
//
// It holds the unit port because converting is Essential's job and never this module's
// (BR-PRICE-UOM-002, PRICE-INV-025). The product port is not held: the line service has already
// resolved the product by the time pricing runs, and asking Inventory the same question twice in
// one write would be a second read for an answer already in hand.
type LinePricer struct {
	uoms itExt.UomExtService
}

func NewLinePricer(uoms itExt.UomExtService) *LinePricer {
	return &LinePricer{uoms: uoms}
}

// PriceLine resolves and stamps the line's vendor price reference.
//
// It writes THREE fields and each answers a different question:
//
//   - vendor_product_price_id — which quote this came from, so a reader can go and look at it;
//   - resolved_unit_price — what that quote said, so an override is visible as a difference;
//   - unit_price — the transacted price, but ONLY when the caller did not state one.
//
// That last condition is the whole of section 29.1. A line carrying an explicit price is a
// negotiated price, and overwriting it with the vendor's list price would undo the negotiation
// every time the line was saved. The resolved figure is recorded beside it instead, which is what
// makes the override auditable rather than invisible.
//
// A failure to resolve is not a failure to save. Pricing is a convenience the buyer may always
// override, and refusing the line because Essential was slow or a vendor has no quotes on file
// would stop somebody typing an order they are perfectly entitled to type.
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
		// An order with no vendor yet is an ordinary draft state, and a line with no product is a
		// free-text charge. Neither can be priced from a vendor product price list, and neither is
		// a fault.
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
		// Explicitly NOT an error and explicitly not a fallback. The line keeps its typed price and
		// records no reference, which is the honest state: nobody has quoted this.
		return nil
	}

	line[models.PurchaseOrderLineFieldVendorProductPriceId] = resolved.VendorProductPriceId
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = resolved.UnitPrice
	if !hasExplicitPrice(line) {
		line[models.PurchaseOrderLineFieldUnitPrice] = resolved.UnitPrice
	}
	return nil
}

// loadCandidates reads this vendor's live quotes for this product and stamps each window verdict
// against priceAt.
//
// The verdict is computed here rather than in the resolver because it needs a clock, and a pure
// function that read the time would answer differently on two runs with the same input. Evaluating
// it once per line also means every candidate is judged against the SAME instant — reading the
// clock per row would let a line saved across midnight rank two quotes by two different dates.
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

// buildRequest expresses the ordered quantity in every unit the candidates quote in.
//
// Once per DISTINCT unit, not once per candidate: a vendor with five quantity breaks in cases would
// otherwise cost five identical conversions. A unit that fails to convert is simply left out of the
// map, and vendorpricing skips the candidates quoting in it — the alternative, a zero, would make
// every break in that unit look reachable.
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
// rather than as an error.
//
// A cross-category refusal is Essential answering correctly, not failing: a quote per kilogram
// simply does not apply to a line counted in litres. Treating it as an error would refuse to save
// the line over a quote nobody asked to use.
func (this *LinePricer) convert(
	ctx corectx.Context, quantity decimal.Decimal, sourceUomId, targetUomId string,
) (decimal.Decimal, bool, error) {
	if targetUomId == sourceUomId || sourceUomId == "" {
		// No unit on the line means the quantity is a bare number, and the only sensible reading is
		// that it is already in whatever the quote is per.
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

// windowCovers reports whether a quote's commercial window is open at the given instant.
//
// Both bounds are optional and an absent bound is open-ended: a quote with no valid_to is a
// standing offer, not an expired one. The bounds are INCLUSIVE, because a quote valid "until the
// 31st" is one a buyer expects to be able to use on the 31st.
//
// It reads through timeOf rather than DynamicFields.GetModelDateTime, which returns nil for a shape
// it does not recognise — including a *ModelDateTime, which is exactly what the model layer's own
// setter produces. Nil means "no bound" here, so an unrecognised value would not fail loudly: it
// would silently turn an expired quote into a standing offer and price an order at a price that had
// been withdrawn.
func windowCovers(row dmodel.DynamicFields, at time.Time) bool {
	if from, ok := timeOf(row, models.VendorProductPriceFieldValidFrom); ok && at.Before(from) {
		return false
	}
	if to, ok := timeOf(row, models.VendorProductPriceFieldValidTo); ok && at.After(to) {
		return false
	}
	return true
}

// hasExplicitPrice reports whether the caller stated a price of its own.
//
// PRESENCE is what distinguishes "the client sent a price" from "the client left it to us" — the
// value cannot, because zero is a legitimate price for a free-of-charge line and treating it as
// absent would silently overwrite one.
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
