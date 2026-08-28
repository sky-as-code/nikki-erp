package services

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Repricing a draft order (BR 70, SALES-011).
//
// **The whole engine re-runs after every single line change.** There is no incremental path, and BR
// 70 is explicit about why: adding or removing one line can trigger a conditional price, make a
// voucher newly eligible or newly not, change which combo applies, or cross a minimum-bill
// threshold. A cheaper update that adjusted only the changed line would get all four of those wrong,
// and would get them wrong silently.
//
// So the adjustment set is replaced WHOLESALE rather than amended. That is also what keeps the
// price explanation (SALES-021) honest: the stored chain is always the chain that produced the
// stored numbers, never a chain plus some leftovers from a previous basket.

// RepriceResult is what a reprice concluded.
type RepriceResult struct {
	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal

	// LineCount is how many lines were priced, so a caller can tell an empty draft from a failure.
	LineCount int
}

// RepriceOrder recomputes an order from its stored lines and writes the result back.
//
// Every write happens in one transaction: the lines, the adjustments and the order totals move
// together or not at all. A partial reprice would leave an order whose totals disagree with its own
// lines, which is worse than one that was never repriced — the second is visibly stale, the first is
// invisibly wrong.
//
// Refuses a non-draft order. Confirmation is the line: after it the numbers are what the business
// promised the customer (BR 11).
// basisSvc re-reads the product's base price and cost. It may be nil: see buildPricingInput for
// why an unavailable Inventory falls back to the stored price rather than refusing to reprice.
func RepriceOrder(
	ctx corectx.Context, orderId string, taxSvc itExt.TaxCalculationExtService, policy SalesPolicy,
	basisSvc itExt.ProductPricingBasisExtService,
) (*RepriceResult, *ft.ClientErrors, error) {
	orderRecord, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if orderRecord == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	order := models.NewSalesOrderFrom(orderRecord)
	if !order.IsEditable() {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("status", "sales_order.not_editable",
			"only a draft sales order may be repriced; this one is '"+
				stringOf(orderRecord, models.SalesOrderFieldStatus)+"'"))
		return nil, vErrs, nil
	}

	input, err := buildPricingInput(ctx, orderRecord, policy, basisSvc)
	if err != nil {
		return nil, nil, err
	}

	// Tax is resolved BEFORE the engine runs and handed in, because the engine is pure (D-13) and
	// must not reach into Accounting. A tax that cannot be determined refuses the reprice rather
	// than being read as zero — see D-41.
	basketTax, vErrs, err := resolveTaxForInput(ctx, taxSvc, orderRecord, input, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	input.Context.TaxByLineKey = basketTax.ByLineKey

	result := pricing.Calculate(input)

	if err := writeRepricedOrder(ctx, orderId, orderRecord, result, basketTax); err != nil {
		return nil, nil, err
	}

	return &RepriceResult{
		Subtotal:      result.Subtotal,
		DiscountTotal: result.DiscountTotal,
		TaxTotal:      result.TaxTotal,
		GrandTotal:    result.GrandTotal,
		LineCount:     len(result.Lines),
	}, nil, nil
}

// buildPricingInput reads the order's stored lines into the engine's input.
//
// The stored line is the source of truth for what is IN the basket — which variant, how many, in
// what unit. Prices are deliberately not taken from it: the base price is re-read from the product
// on every reprice, because a draft left open across a price change must reflect the new price
// rather than the one it was created under.
//
// That re-read is what basisSvc is for. Until it existed this function took base_unit_price from
// the stored line, which made a reprice reproduce the price it already had whenever no pricelist
// rule matched — the fallback path silently never moved.
//
// A nil port FAILS SOFT, falling back to the stored line's price. Sales must keep pricing when
// Inventory is slow or mid-restart, and the stored price is the last known good answer rather than
// a guess. The alternative — refusing to reprice — would block a till over another module's
// availability, and the CR asks for repricing to be explicit, not fragile.
func buildPricingInput(
	ctx corectx.Context, orderRecord dmodel.DynamicFields, policy SalesPolicy,
	basisSvc itExt.ProductPricingBasisExtService,
) (pricing.Input, error) {
	orderId := stringOf(orderRecord, models.SalesOrderFieldId)

	lineRecords, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return pricing.Input{}, err
	}

	lines := make([]pricing.LineInput, 0, len(lineRecords))
	for _, record := range lineRecords {
		// A giveaway line is an OUTPUT of pricing, not an input to it. Feeding one back in would
		// make the engine give away a second free item on every reprice, compounding each time.
		if stringOf(record, models.SalesOrderLineFieldLineType) ==
			string(models.SalesOrderLineTypePromotionReward) {
			continue
		}

		lines = append(lines, pricing.LineInput{
			Key:                stringOf(record, models.SalesOrderLineFieldId),
			LineNumber:         int32Of(record, models.SalesOrderLineFieldLineNumber),
			ProductVariantId:   stringOf(record, models.SalesOrderLineFieldProductVariantId),
			UomId:              stringOf(record, models.SalesOrderLineFieldUomId),
			Quantity:           decimalOf(record, models.SalesOrderLineFieldOrderedQuantity),
			CatalogueUnitPrice: decimalOf(record, models.SalesOrderLineFieldBaseUnitPrice),
			ProductCode:        stringOf(record, models.SalesOrderLineFieldProductCodeSnapshot),
			ProductName:        stringOf(record, models.SalesOrderLineFieldProductNameSnapshot),
			ComboId:            stringOf(record, models.SalesOrderLineFieldSalesComboId),
		})
	}

	// One batch read for the whole basket, then each line takes its own row. Done after the loop
	// rather than inside it so that a twenty-line order costs one round trip instead of twenty.
	applyPricingBasis(ctx, lines, basisSvc)

	// Line-number order, so a reprice of an unchanged basket produces an identical chain. The
	// allocator breaks residual ties on line number (D-04), so shuffled input would move sub-unit
	// amounts between lines and BR 13's "same input, same output" would fail on a reread.
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].LineNumber < lines[j].LineNumber
	})

	// The operator overrides are loaded as ENGINE INPUT, not left as adjustments to preserve.
	// replaceAdjustments below deletes the whole chain and rewrites it from engine output, so an
	// override that was not fed back in would be erased by this very function - and confirm reprices
	// unconditionally, so it would be erased before the sale completed (BR 87.4, SALES-039).
	manual, err := loadManualDiscounts(ctx, orderId)
	if err != nil {
		return pricing.Input{}, err
	}

	return pricing.Input{
		Lines:           lines,
		ManualDiscounts: manual,
		Context: pricing.Context{
			CurrencyScale: policy.RoundingScale,
		},
	}, nil
}

// loadManualDiscounts reads the order's stored overrides in a stable order.
//
// Sorted by id, for the same reason the lines are sorted by line number: the engine applies them in
// the order given, each capped against what the previous ones left, so a shuffled read would move
// money between overrides and break BR 13's "same input, same output" on a mere reread.
func loadManualDiscounts(
	ctx corectx.Context, orderId string,
) ([]pricing.ManualDiscountInput, error) {
	records, err := searchBy(ctx,
		models.SalesManualDiscountSchemaName,
		models.SalesManualDiscountFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	discounts := make([]pricing.ManualDiscountInput, 0, len(records))
	for _, record := range records {
		discounts = append(discounts, pricing.ManualDiscountInput{
			Id:      stringOf(record, models.SalesManualDiscountFieldId),
			LineKey: stringOf(record, models.SalesManualDiscountFieldOrderLineId),
			Amount:  decimalOf(record, models.SalesManualDiscountFieldAmount),
			Reason:  stringOf(record, models.SalesManualDiscountFieldReason),
		})
	}
	sort.SliceStable(discounts, func(i, j int) bool {
		return discounts[i].Id < discounts[j].Id
	})
	return discounts, nil
}

// writeRepricedOrder replaces the lines, the adjustments and the totals, in one transaction.
func writeRepricedOrder(
	ctx corectx.Context,
	orderId string,
	orderRecord dmodel.DynamicFields,
	result pricing.Result,
	basketTax *BasketTax,
) error {
	return withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		if err := writeLineResults(tranxCtx, result.Lines); err != nil {
			return err
		}
		if err := replaceAdjustments(tranxCtx, orderId, result.Adjustments); err != nil {
			return err
		}
		return writeOrderTotals(tranxCtx, orderId, orderRecord, result, basketTax)
	})
}

// writeLineResults pushes the computed figures back onto each line.
func writeLineResults(ctx corectx.Context, lines []pricing.LineResult) error {
	engine, err := engineFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}

	for _, line := range lines {
		update := dmodel.DynamicFields{
			models.SalesOrderLineFieldId:                 line.Key,
			models.SalesOrderLineFieldEffectiveUnitPrice: line.EffectiveUnitPrice,
			models.SalesOrderLineFieldGrossAmount:        line.GrossAmount,
			models.SalesOrderLineFieldDiscountAmount:     line.DiscountAmount,
			models.SalesOrderLineFieldNetAmount:          line.NetAmount,
			models.SalesOrderLineFieldTaxRateSnapshot:    line.TaxRateSnapshot,
			models.SalesOrderLineFieldTaxAmount:          line.TaxAmount,
			models.SalesOrderLineFieldFinalAmount:        line.FinalAmount,
			models.SalesOrderLineFieldPricingSource:      line.PricingSource,
		}
		if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
			return err
		}
	}
	return nil
}

// replaceAdjustments deletes the order's adjustment chain and writes the new one.
//
// Replaced rather than amended, because sequence is unique per order and the new chain renumbers
// from one. Amending would either collide on the unique index or leave a chain that no longer
// replays to the stored totals — and the price explanation reads exactly this table.
func replaceAdjustments(
	ctx corectx.Context, orderId string, adjustments []pricing.Adjustment,
) error {
	engine, err := engineFor(models.SalesOrderAdjustmentSchemaName)
	if err != nil {
		return err
	}
	repo := engine.ResourceRepository()

	existing, err := searchBy(ctx,
		models.SalesOrderAdjustmentSchemaName,
		models.SalesOrderAdjustmentFieldSalesOrderId, orderId)
	if err != nil {
		return err
	}
	for _, record := range existing {
		keys := dmodel.DynamicFields{
			models.SalesOrderAdjustmentFieldId: stringOf(record, models.SalesOrderAdjustmentFieldId),
		}
		if _, err := repo.DeleteOne(ctx, keys); err != nil {
			return err
		}
	}

	orgId := ""
	if len(existing) > 0 {
		orgId = stringOf(existing[0], basemodel.FieldOrgId)
	}

	for _, adjustment := range adjustments {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		fields := dmodel.DynamicFields{
			models.SalesOrderAdjustmentFieldId:               string(*id),
			models.SalesOrderAdjustmentFieldSalesOrderId:     orderId,
			models.SalesOrderAdjustmentFieldSequence:         adjustment.Sequence,
			models.SalesOrderAdjustmentFieldAdjustmentType:   adjustment.Type,
			models.SalesOrderAdjustmentFieldBaseAmount:       adjustment.BaseAmount,
			models.SalesOrderAdjustmentFieldAdjustmentAmount: adjustment.Amount,
		}
		if adjustment.LineKey != "" {
			fields[models.SalesOrderAdjustmentFieldSalesOrderLineId] = adjustment.LineKey
		}
		if adjustment.SourceType != "" {
			fields[models.SalesOrderAdjustmentFieldSourceType] = adjustment.SourceType
		}
		if adjustment.SourceId != "" {
			fields[models.SalesOrderAdjustmentFieldSourceId] = adjustment.SourceId
		}
		if adjustment.Description != "" {
			fields[models.SalesOrderAdjustmentFieldDescription] = adjustment.Description
		}
		if orgId != "" {
			fields[basemodel.FieldOrgId] = orgId
		}
		if _, err := repo.Insert(ctx, fields); err != nil {
			return err
		}
	}
	return nil
}

// writeOrderTotals stamps the recomputed totals and the tax snapshot onto the order.
func writeOrderTotals(
	ctx corectx.Context,
	orderId string,
	orderRecord dmodel.DynamicFields,
	result pricing.Result,
	basketTax *BasketTax,
) error {
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.SalesOrderFieldId:            orderId,
		models.SalesOrderFieldSubtotal:      result.Subtotal,
		models.SalesOrderFieldDiscountTotal: result.DiscountTotal,
		models.SalesOrderFieldTaxTotal:      result.TaxTotal,
		models.SalesOrderFieldGrandTotal:    result.GrandTotal,
	}

	// The snapshot is stored only when Accounting actually produced one. Writing an empty snapshot
	// over a real one would destroy the only surviving record of how a sale was taxed
	// (BR-TAX-ESS-030).
	if basketTax != nil && basketTax.Snapshot.SchemaVersion != "" {
		update[models.SalesOrderFieldTaxSnapshot] = basketTax.Snapshot
	}

	_, err = engine.ResourceRepository().Update(ctx, update)
	return err
}

// resolveTaxForInput asks Accounting for the tax on the basket about to be priced.
//
// It runs on the PRE-pricing line amounts rather than the post-pricing ones, and that is a real
// approximation worth naming: the taxable base should be each line's net after discounts, but the
// discounts are what the engine is about to compute. Resolving tax first is nonetheless correct for
// the tax-INCLUSIVE case this module sells under (D-41): with an inclusive price the tax is already
// inside the number, so the rate is what matters and the base moves with the discount automatically.
//
// It would be wrong under tax-exclusive pricing, where tax is added on top of a base the discount
// changes. Accounting decides which applies from its own price_inclusion_mode, so if a deployment
// ever configures an exclusive tax this needs a second pass: price, then tax, then re-total. Flagged
// rather than guessed at, because today every seeded tax is `inherit` against an included price mode.
func resolveTaxForInput(
	ctx corectx.Context,
	taxSvc itExt.TaxCalculationExtService,
	orderRecord dmodel.DynamicFields,
	input pricing.Input,
	policy SalesPolicy,
) (*BasketTax, *ft.ClientErrors, error) {
	// The lines the tax service sees carry their pre-discount amounts, since the engine has not run.
	lines := make([]pricing.LineResult, 0, len(input.Lines))
	for _, line := range input.Lines {
		gross := line.CatalogueUnitPrice.Mul(line.Quantity)
		lines = append(lines, pricing.LineResult{
			Key:              line.Key,
			LineNumber:       line.LineNumber,
			ProductVariantId: line.ProductVariantId,
			UomId:            line.UomId,
			Quantity:         line.Quantity,
			GrossAmount:      gross,
			NetAmount:        gross,
		})
	}

	return ResolveBasketTax(ctx, taxSvc, TaxRequestContext{
		OrgId:               stringOf(orderRecord, basemodel.FieldOrgId),
		TaxDate:             taxDateOf(orderRecord),
		CurrencyCode:        stringOf(orderRecord, models.SalesOrderFieldCurrencyCode),
		TaxCode:             policy.DefaultSalesTaxCode,
		PriceMode:           itExt.PriceInclusionIncluded,
		BusinessChannelCode: stringOf(orderRecord, models.SalesOrderFieldSalesChannelId),
		OutletReference:     stringOf(orderRecord, models.SalesOrderFieldSalesPointId),
	}, lines)
}

// taxDateOf is the date the sale legally occurred, formatted YYYY-MM-DD.
//
// A confirmed order is taxed as of its confirmation; a draft as of today, because it has not legally
// happened yet. BR-TAX-ESS-SUP-020 forbids Accounting from defaulting this from its own clock, so the
// caller must always supply one — which is why this never returns empty.
func taxDateOf(orderRecord dmodel.DynamicFields) string {
	if confirmed := dateTimeOf(orderRecord, models.SalesOrderFieldConfirmedAt); confirmed != nil {
		return confirmed.GoTime().Format("2006-01-02")
	}
	return time.Now().UTC().Format("2006-01-02")
}

// applyPricingBasis fills each line's product-derived pricing inputs, in place.
//
// It reports nothing. Every failure here is recoverable by leaving the line as it was: the stored
// base price stands, the line still prices, and the order still totals. Returning an error would
// turn a slow read in another module into a refused sale, which is a worse answer than pricing at
// the price the line already carried.
//
// A variant missing from the answer is left alone for the same reason. It means the product was
// deleted or is not visible, and neither is something a reprice can fix — but the line is still in
// the basket and still has to total.
func applyPricingBasis(
	ctx corectx.Context, lines []pricing.LineInput, basisSvc itExt.ProductPricingBasisExtService,
) {
	if basisSvc == nil || len(lines) == 0 {
		return
	}

	variantIds := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.ProductVariantId != "" {
			variantIds = append(variantIds, line.ProductVariantId)
		}
	}
	if len(variantIds) == 0 {
		return
	}

	result, err := basisSvc.GetPricingBasis(ctx, itExt.GetPricingBasisQuery{
		ProductVariantIds: variantIds,
	})
	if err != nil || result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return
	}

	for index := range lines {
		basis, found := result.Data.Bases[lines[index].ProductVariantId]
		if !found {
			continue
		}
		lines[index].ProductTemplateId = basis.ProductTemplateId
		lines[index].CategoryPath = basis.CategoryPath

		// An unparseable or absent base price leaves the stored one standing. Zero is a real price
		// — a giveaway — so it must not be conjured out of a failed parse.
		if price, ok := parseDecimal(basis.EffectiveBaseSalesPrice); ok {
			lines[index].CatalogueUnitPrice = price
		}
		if basis.HasCost {
			if cost, ok := parseDecimal(basis.Cost); ok {
				lines[index].UnitCost, lines[index].HasCost = cost, true
			}
		}
	}
}

// parseDecimal reads a decimal that crossed a module boundary as a string.
//
// Decimals travel as strings between modules on purpose: a price carried as float64 loses exactly
// the precision that must not be lost. The bool distinguishes "absent or malformed" from a
// legitimate zero.
func parseDecimal(text string) (decimal.Decimal, bool) {
	if text == "" {
		return decimal.Zero, false
	}
	value, err := decimal.NewFromString(text)
	if err != nil {
		return decimal.Zero, false
	}
	return value, true
}
