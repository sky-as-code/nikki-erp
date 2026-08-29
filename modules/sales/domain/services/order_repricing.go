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

// Repricing a draft order. The whole engine re-runs on every line change and the adjustment set is
// replaced wholesale rather than amended: one added line can change conditional prices, voucher
// eligibility, combo choice or a minimum-bill threshold, and the stored chain must always be the
// chain that produced the stored numbers.

// RepriceResult is what a reprice concluded.
type RepriceResult struct {
	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal

	// LineCount lets a caller tell an empty draft from a failure.
	LineCount int
}

// RepriceOrder recomputes an order from its stored lines and writes the result back.
//
// Lines, adjustments and totals move in one transaction, since a partial reprice leaves totals that
// disagree with the lines — invisibly wrong rather than visibly stale. Refuses a non-draft order,
// because after confirmation the numbers are what was promised the customer. basisSvc may be nil;
// see buildPricingInput.
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

	// Tax is resolved before the engine runs and handed in, because the engine is pure and must not
	// reach into Accounting. A tax that cannot be determined refuses the reprice rather than reading
	// as zero.
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
// The stored line says what is in the basket but not at what price: the base price is re-read from
// the product via basisSvc on every reprice, so a draft left open across a price change reflects the
// new price. A nil or failing basisSvc falls back to the stored line's price rather than refusing to
// reprice, so another module's availability cannot block a till.
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
		// A giveaway line is an output of pricing; feeding one back in would give away another free
		// item on every reprice, compounding each time.
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

	// One batch read after the loop, so a twenty-line order costs one round trip instead of twenty.
	applyPricingBasis(ctx, lines, basisSvc)

	// Line-number order, so an unchanged basket reprices to an identical chain. The allocator breaks
	// residual ties on line number, so shuffled input would move sub-unit amounts between lines.
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].LineNumber < lines[j].LineNumber
	})

	// Operator overrides are engine input, not adjustments to preserve: replaceAdjustments rewrites
	// the whole chain from engine output, so an override not fed back in would be erased — and
	// confirm reprices unconditionally, so it would be erased before the sale completed.
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

// loadManualDiscounts reads the order's stored overrides sorted by id. The engine applies them in
// the order given, each capped against what the previous ones left, so a shuffled read would move
// money between overrides.
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

// replaceAdjustments deletes the order's adjustment chain and writes the new one. Replaced rather
// than amended, because sequence is unique per order and the new chain renumbers from one; the price
// explanation replays exactly this table.
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

	// Only stored when Accounting produced one: an empty snapshot written over a real one destroys
	// the only surviving record of how a sale was taxed.
	if basketTax != nil && basketTax.Snapshot.SchemaVersion != "" {
		update[models.SalesOrderFieldTaxSnapshot] = basketTax.Snapshot
	}

	_, err = engine.ResourceRepository().Update(ctx, update)
	return err
}

// resolveTaxForInput asks Accounting for the tax on the basket about to be priced, using pre-pricing
// line amounts because the discounts are what the engine is about to compute. That holds only under
// tax-inclusive pricing, where the tax sits inside the number and the base moves with the discount;
// a deployment configuring exclusive tax needs a second pass: price, then tax, then re-total.
func resolveTaxForInput(
	ctx corectx.Context,
	taxSvc itExt.TaxCalculationExtService,
	orderRecord dmodel.DynamicFields,
	input pricing.Input,
	policy SalesPolicy,
) (*BasketTax, *ft.ClientErrors, error) {
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

// taxDateOf is the date the sale legally occurred, formatted YYYY-MM-DD: a confirmed order is taxed
// as of its confirmation, a draft as of today. Accounting may not default this from its own clock,
// so this never returns empty.
func taxDateOf(orderRecord dmodel.DynamicFields) string {
	if confirmed := dateTimeOf(orderRecord, models.SalesOrderFieldConfirmedAt); confirmed != nil {
		return confirmed.GoTime().Format("2006-01-02")
	}
	return time.Now().UTC().Format("2006-01-02")
}

// applyPricingBasis fills each line's product-derived pricing inputs, in place.
//
// It reports nothing on purpose: every failure, including a variant missing from the answer, leaves
// the line's stored base price standing so the order still totals. Erroring would turn a slow read
// in another module into a refused sale.
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

		// An absent or unparseable price leaves the stored one standing: zero is a real price — a
		// giveaway — so it must not come out of a failed parse.
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

// parseDecimal reads a decimal that crossed a module boundary as a string (float64 would lose the
// precision that must not be lost). The bool distinguishes absent or malformed from a legitimate
// zero.
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
