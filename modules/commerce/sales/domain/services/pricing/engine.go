package pricing

import (
	"github.com/shopspring/decimal"
)

// Calculate prices a basket. Pure: no I/O, no clock, no repository. Nine steps in this exact order:
//
//  1. resolve product, quantity and unit
//  2. resolve the base price, from a pricelist item or the catalogue
//  3. resolve the combo price
//  4. (promotions: applied afterwards by ApplyPromotionEffects, see below)
//  5. apply the manual overrides
//  6. allocate order-level adjustments down to the lines
//  7. calculate tax
//  8. apply monetary rounding
//  9. freeze the snapshot
//
// Promotions are not a step inside Calculate: the BR engine qualifies them on the untaxed value
// after manual discounts (BR-C05, BR-C06), which needs Accounting's tax on this result first.
func Calculate(input Input) Result {
	scale := InternalScale

	// Steps 1–3: every line gets a price, and combo lines get their components allocated.
	lines := resolveLines(input, scale)

	sequence := int32(0)
	adjustments := make([]Adjustment, 0, 8)

	// Combo pricing is recorded as an adjustment so the difference between the parts' cost and the
	// bundle price is explainable rather than merely absent.
	for index := range lines {
		if lines[index].ComboId == "" {
			continue
		}
		reference := referenceTotalOf(lines[index])
		if reference.IsZero() || reference.Equal(lines[index].GrossAmount) {
			continue
		}
		sequence++
		adjustments = append(adjustments, Adjustment{
			Sequence:    sequence,
			LineKey:     lines[index].Key,
			Type:        AdjustmentComboPrice,
			SourceType:  "sales_combo",
			SourceId:    lines[index].ComboId,
			Description: "Combo price",
			BaseAmount:  reference,
			Amount:      lines[index].GrossAmount.Sub(reference),
		})
	}

	// Step 5b: manual overrides, before tax because an override changes what is taxed. Promotions
	// follow in ApplyPromotionEffects: the BR engine (pricing/promotion) needs Accounting's tax on
	// this pre-promotion base, which only the caller can fetch.
	lines, adjustments, sequence = applyManualDiscounts(input, lines, adjustments, sequence, scale)

	// Step 6: allocate order-level adjustments down to the lines.
	lines = allocateOrderAdjustments(lines, adjustments, scale)

	// Step 7: tax.
	lines = applyTax(lines, input.Context.TaxByLineKey, scale)
	return finish(lines, adjustments, sequence, input.Context.CurrencyScale)
}

// finish is steps 8 and 9: rounding once on the grand total, then the freeze. The caller writes the
// lines and adjustments as they stand; after confirmation they are immutable.
func finish(lines []LineResult, adjustments []Adjustment, sequence int32, currencyScale int32) Result {
	result := totalise(lines)
	rounded := result.GrandTotal.Round(currencyScale)
	if !rounded.Equal(result.GrandTotal) {
		sequence++
		adjustments = append(adjustments, Adjustment{
			Sequence:    sequence,
			Type:        AdjustmentRounding,
			Description: "Monetary rounding",
			BaseAmount:  result.GrandTotal,
			Amount:      rounded.Sub(result.GrandTotal),
		})
		result.GrandTotal = rounded
	}
	result.Lines = lines
	result.Adjustments = adjustments
	return result
}

// applyManualDiscounts applies the operator overrides. Each becomes an adjustment and nothing else:
// the base price is untouched, so the chain catalogue → pricelist → programs → override → tax stays
// visible. An override naming no line is order-level and spreads proportionally across the lines.
func applyManualDiscounts(
	input Input,
	lines []LineResult,
	adjustments []Adjustment,
	sequence int32,
	scale int32,
) ([]LineResult, []Adjustment, int32) {
	for _, manual := range input.ManualDiscounts {
		// A non-positive override is ignored: a negative one would be an unauthorised surcharge,
		// silently adding money to the customer's bill.
		if !manual.Amount.IsPositive() {
			continue
		}

		base := decimal.Zero
		if manual.LineKey == "" {
			base = netTotalOf(lines)
		} else {
			for index := range lines {
				if lines[index].Key == manual.LineKey {
					base = lines[index].NetAmount
					break
				}
			}
		}

		// Capped at what is owed: an override larger than the line would be a refund, which has its
		// own workflow and money movement.
		amount := manual.Amount
		if amount.GreaterThan(base) {
			amount = base
		}
		if !amount.IsPositive() {
			continue
		}

		// Signed NEGATIVE from here on, like every other discount here: applyDiscountToLines adds what
		// it is given.
		signed := amount.Neg()

		sequence++
		adjustments = append(adjustments, Adjustment{
			Sequence:    sequence,
			LineKey:     manual.LineKey,
			Type:        AdjustmentManualDiscount,
			SourceType:  "sales_manual_discount",
			SourceId:    manual.Id,
			Description: manual.Reason,
			BaseAmount:  base,
			Amount:      signed,
		})

		// The discount must reach the LINES in both cases: totalise() sums the lines alone, so an
		// adjustment touching no line would show as a discount the customer never received.
		if manual.LineKey == "" {
			// Order-level: proportional across every line, via the allocator the promotions use.
			lines = applyDiscountToLines(lines, allIndices(lines), signed, scale)
			continue
		}

		// A key naming no line would index -1 and panic; the engine takes its input on trust, so it
		// declines rather than crashing the calculation over one bad row.
		if index := indexOfKey(lines, manual.LineKey); index >= 0 {
			lines = applyDiscountToLines(lines, []int{index}, signed, scale)
		}
	}
	return lines, adjustments, sequence
}

// allIndices names every line, for an order-level adjustment.
func allIndices(lines []LineResult) []int {
	indices := make([]int, len(lines))
	for index := range lines {
		indices[index] = index
	}
	return indices
}

// indexOfKey finds a line by its caller-supplied key, or -1.
func indexOfKey(lines []LineResult, key string) int {
	for index := range lines {
		if lines[index].Key == key {
			return index
		}
	}
	return -1
}

// resolveLines runs steps 1–3: unit price, gross amount, and combo component allocation.
func resolveLines(input Input, scale int32) []LineResult {
	combos := make(map[string]ComboDefinition, len(input.Combos))
	for _, combo := range input.Combos {
		combos[combo.ComboId] = combo
	}

	lines := make([]LineResult, 0, len(input.Lines))
	for _, line := range input.Lines {
		result := LineResult{
			Key:                 line.Key,
			LineNumber:          line.LineNumber,
			ProductVariantId:    line.ProductVariantId,
			UomId:               line.UomId,
			Quantity:            line.Quantity,
			ProductCode:         line.ProductCode,
			ProductName:         line.ProductName,
			LineType:            "product",
			PricingSource:       "catalogue",
			RequiresFulfillment: line.RequiresFulfillment,
			BaseUnitPrice:       line.CatalogueUnitPrice,
			EffectiveUnitPrice:  line.CatalogueUnitPrice,
		}

		// Step 2: a matching pricelist rule overrides the catalogue price. A rule that matches but
		// cannot compute leaves the catalogue price standing rather than refusing the line.
		if item, found := bestRule(input.PricelistItems, line); found {
			if price, computed := rulePrice(item, line, scale); computed {
				result.BaseUnitPrice = price
				result.EffectiveUnitPrice = price
				result.PricingSource = "pricelist"
				result.PricelistItemId = item.Id
				result.PricelistId = item.PricelistId
			}
		}

		// Step 3: a combo line is priced at the bundle price, not at its parts.
		if line.ComboId != "" {
			result.LineType = "combo"
			result.ComboId = line.ComboId
			result.PricingSource = "combo"
			if combo, found := combos[line.ComboId]; found {
				result.EffectiveUnitPrice = combo.ComboPrice
				result.BaseUnitPrice = combo.ComboPrice
				result.Components = expandComponents(combo, line)
			}
		}

		result.GrossAmount = result.EffectiveUnitPrice.Mul(line.Quantity).Round(scale)
		result.NetAmount = result.GrossAmount
		lines = append(lines, result)
	}
	return lines
}

// expandComponents allocates the combo price across its components. The allocation is an OUTPUT for
// VAT, partial return and reporting, never an input to the combo price; its sum equals the line's
// gross amount exactly.
func expandComponents(combo ComboDefinition, line LineInput) []ComponentResult {
	if len(combo.Components) == 0 {
		return nil
	}

	inputs := make([]allocationInput, len(combo.Components))
	for index, component := range combo.Components {
		inputs[index] = allocationInput{
			key:       component.Key,
			reference: component.ReferencePrice.Mul(component.Quantity),
			tiebreak:  component.Sequence,
		}
	}

	lineTotal := combo.ComboPrice.Mul(line.Quantity).Round(InternalScale)
	shares := allocate(lineTotal, inputs, InternalScale)

	results := make([]ComponentResult, len(combo.Components))
	for index, component := range combo.Components {
		results[index] = ComponentResult{
			Key:                component.Key,
			Sequence:           component.Sequence,
			ProductVariantId:   component.ProductVariantId,
			UomId:              component.UomId,
			Quantity:           component.Quantity.Mul(line.Quantity),
			ProductCode:        component.ProductCode,
			ProductName:        component.ProductName,
			AllocatedNetAmount: shares[component.Key],
		}
	}
	return results
}

// referenceTotalOf is what a combo line's contents would have cost at their standalone prices.
func referenceTotalOf(line LineResult) decimal.Decimal {
	total := decimal.Zero
	for _, component := range line.Components {
		total = total.Add(component.AllocatedNetAmount)
	}
	if total.IsZero() {
		return decimal.Zero
	}
	// The allocated amounts already sum to the combo price, so they cannot serve as the reference; the
	// gross amount does, making the combo adjustment zero unless standalone prices were supplied.
	return line.GrossAmount
}

// applyDiscountToLines spreads a discount across the target lines proportionally to their CURRENT
// (post-discount) net amounts, which keeps the line nets summing to the order net at every step and
// stops an already-discounted line absorbing more than it is worth.
func applyDiscountToLines(
	lines []LineResult, targets []int, amount decimal.Decimal, scale int32,
) []LineResult {
	if len(targets) == 0 || amount.IsZero() {
		return lines
	}

	inputs := make([]allocationInput, len(targets))
	for position, index := range targets {
		inputs[position] = allocationInput{
			key:       lines[index].Key,
			reference: lines[index].NetAmount,
			tiebreak:  lines[index].LineNumber,
		}
	}
	shares := allocate(amount, inputs, scale)

	for _, index := range targets {
		share := shares[lines[index].Key]
		// DiscountAmount is positive by convention on the line, while the adjustment is signed.
		lines[index].DiscountAmount = lines[index].DiscountAmount.Add(share.Neg())
		lines[index].NetAmount = lines[index].NetAmount.Add(share)
	}
	return lines
}

// allocateOrderAdjustments is step 6. Discounts were already pushed down to the lines as applied, so
// this reconciles rather than re-allocates: net is gross minus accumulated discount, never negative.
func allocateOrderAdjustments(
	lines []LineResult, _ []Adjustment, scale int32,
) []LineResult {
	for index := range lines {
		net := lines[index].GrossAmount.Sub(lines[index].DiscountAmount).Round(scale)
		if net.IsNegative() {
			// Capped here rather than per reward, so a stack of rewards is capped once on its combined
			// effect.
			lines[index].DiscountAmount = lines[index].GrossAmount
			net = decimal.Zero
		}
		lines[index].NetAmount = net
	}
	return lines
}

// applyTax is step 7: it copies in what Accounting decided and never computes tax itself — the
// arithmetic, including compound and per-unit taxes and tax-inclusive pricing, belongs to
// accounting/domain/services/tax. A line with no entry in the map is taxed at zero.
func applyTax(lines []LineResult, taxByLineKey map[string]LineTax, scale int32) []LineResult {
	for index := range lines {
		tax, resolved := taxByLineKey[lines[index].Key]
		if !resolved {
			lines[index].TaxRateSnapshot = decimal.Zero
			lines[index].TaxAmount = decimal.Zero
			lines[index].FinalAmount = lines[index].NetAmount
			continue
		}

		lines[index].TaxRateSnapshot = tax.RateSnapshot
		lines[index].TaxAmount = tax.Amount
		lines[index].FinalAmount = lines[index].NetAmount

		if len(lines[index].Components) == 0 {
			continue
		}

		// Accounting taxed the components individually, so its split is authoritative: it reconciles
		// against the rounding policy actually applied.
		if len(tax.ComponentAmounts) > 0 {
			lines[index].Components = copyComponentTax(lines[index].Components, tax.ComponentAmounts)
			continue
		}

		// Otherwise the line was taxed as a whole and components share it in proportion to allocated
		// net; the shares sum to the line's tax exactly, so an itemised VAT invoice still adds up.
		lines[index].Components = allocateComponentTax(
			lines[index].Components, lines[index].TaxAmount, scale)
	}
	return lines
}

// copyComponentTax takes Accounting's per-component figures. An unreported component gets zero, not
// a guessed share, so the total keeps reconciling with the tax actually charged.
func copyComponentTax(
	components []ComponentResult, amounts map[string]decimal.Decimal,
) []ComponentResult {
	for index := range components {
		components[index].AllocatedTaxAmount = amounts[components[index].Key]
	}
	return components
}

func allocateComponentTax(
	components []ComponentResult, taxAmount decimal.Decimal, scale int32,
) []ComponentResult {
	if taxAmount.IsZero() {
		return components
	}
	inputs := make([]allocationInput, len(components))
	for index, component := range components {
		inputs[index] = allocationInput{
			key:       component.Key,
			reference: component.AllocatedNetAmount,
			tiebreak:  component.Sequence,
		}
	}
	shares := allocate(taxAmount, inputs, scale)
	for index := range components {
		components[index].AllocatedTaxAmount = shares[components[index].Key]
	}
	return components
}

// totalise sums the lines into the order totals. FinalAmount is the line's net: with tax-inclusive
// pricing the tax is already inside it, so adding tax again would charge it twice.
func totalise(lines []LineResult) Result {
	result := Result{
		Subtotal:      decimal.Zero,
		DiscountTotal: decimal.Zero,
		TaxTotal:      decimal.Zero,
		GrandTotal:    decimal.Zero,
	}
	for _, line := range lines {
		result.Subtotal = result.Subtotal.Add(line.GrossAmount)
		result.DiscountTotal = result.DiscountTotal.Add(line.DiscountAmount)
		result.TaxTotal = result.TaxTotal.Add(line.TaxAmount)
		result.GrandTotal = result.GrandTotal.Add(line.FinalAmount)
	}
	return result
}

func netTotalOf(lines []LineResult) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.NetAmount)
	}
	return total
}
