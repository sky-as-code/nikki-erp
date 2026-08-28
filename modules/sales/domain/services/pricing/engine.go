package pricing

import (
	"sort"

	"github.com/shopspring/decimal"
)

// Calculate prices a basket. Pure: no I/O, no clock, no repository (D-13).
//
// The nine steps of BR §13, in this exact order:
//
//  1. resolve product, quantity and unit
//  2. resolve the base price, from a pricelist item or the catalogue
//  3. resolve the combo price
//  4. apply automatic conditional pricing
//  5. apply voucher programs
//  6. allocate order-level adjustments down to the lines
//  7. calculate tax
//  8. apply monetary rounding
//  9. freeze the snapshot
//
// Steps 4 and 5 are one loop rather than two. The programs arrive already ordered by the caller's
// conflict resolution, and BR §21–22 forbid separate engines for promotions and vouchers — running
// them as two passes would be exactly the separate engines the requirement rules out, and the
// second pass would silently take precedence over the first whatever the priorities said.
//
// A program is evaluated against the state BEFORE any program was applied, not against the running
// total. BR §29 does not address this, so the choice is documented here: evaluating against the
// running total would make a program's eligibility depend on which other programs happened to be
// applied first, so the same basket could qualify for a discount or not depending on ordering that
// the customer cannot see. The engine takes the stable reading.
func Calculate(input Input) Result {
	scale := InternalScale

	// Steps 1–3: every line gets a price, and combo lines get their components allocated.
	lines := resolveLines(input, scale)

	sequence := int32(0)
	adjustments := make([]Adjustment, 0, 8)

	// Combo pricing is recorded as an adjustment so the difference between what the parts would
	// have cost and what the bundle costs is explainable rather than merely absent.
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

	// Steps 4–5: apply the programs the caller resolved, in the order given.
	lines, adjustments, sequence = applyPrograms(input, lines, adjustments, sequence, scale)

	// Step 5b: manual overrides (BR 87.4). AFTER the programs, because a human decision is made in
	// light of what the automatic pricing already produced - an operator granting goodwill on a
	// basket has seen the discounted number, not the catalogue one. Before tax, because an override
	// changes what is being taxed.
	lines, adjustments, sequence = applyManualDiscounts(input, lines, adjustments, sequence, scale)

	// Step 6: allocate order-level adjustments down to the lines.
	lines = allocateOrderAdjustments(lines, adjustments, scale)

	// Step 7: tax.
	lines = applyTax(lines, input.Context.TaxByLineKey, scale)

	// Step 8: rounding, the last money step before the freeze. Applied once, to the grand total.
	result := totalise(lines)
	rounded := result.GrandTotal.Round(input.Context.CurrencyScale)
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

	// Step 9: freeze. The lines and adjustments are returned as they stand; the caller writes them
	// and, after confirmation, they are immutable (BR §11).
	result.Lines = lines
	result.Adjustments = adjustments
	return result
}

// applyManualDiscounts applies the operator overrides (BR 87.4).
//
// Each one becomes an adjustment and nothing else. The base price is untouched, which is what BR 87.4
// requires and what keeps the price explanation readable: the chain still runs catalogue price →
// pricelist → programs → override → tax, with every link visible.
//
// An override naming a line is applied to that line. One naming none is an order-level adjustment,
// spread across the lines by step 6 like any other - so a goodwill discount on a basket lands
// proportionally rather than arbitrarily on the first line.
func applyManualDiscounts(
	input Input,
	lines []LineResult,
	adjustments []Adjustment,
	sequence int32,
	scale int32,
) ([]LineResult, []Adjustment, int32) {
	for _, manual := range input.ManualDiscounts {
		// A non-positive override is ignored rather than applied. Zero changes nothing, and a
		// negative one is a surcharge, which BR 87.4 does not authorise - and silently adding money
		// to a customer's bill is the worst available failure here.
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

		// Capped at what is actually owed. An override larger than the line cannot make the
		// customer owe a negative amount - that is a refund, which has its own workflow and its own
		// money movement, not a discount that ran past the end.
		amount := manual.Amount
		if amount.GreaterThan(base) {
			amount = base
		}
		if !amount.IsPositive() {
			continue
		}

		// Signed NEGATIVE from here on, matching every other discount in this engine: the
		// adjustment records a reduction, and applyDiscountToLines adds what it is given.
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

		// The discount is spread onto the LINES in both cases, and that is not optional: step 6
		// ignores the adjustment list and totalise() sums the lines alone, so an adjustment that
		// touched no line would appear on the paperwork as a discount the customer never received.
		if manual.LineKey == "" {
			// Order-level: proportional across every line, by the same allocator the promotions
			// use, so a goodwill discount on a basket lands the way every other order-level
			// discount does rather than arbitrarily on the first line.
			lines = applyDiscountToLines(lines, allIndices(lines), signed, scale)
			continue
		}

		// A key naming no line would index -1 and panic. It cannot normally happen - the caller
		// validates the line exists before storing the override - but the engine is pure and takes
		// its input on trust, so it declines rather than crashing the whole calculation over one
		// bad row.
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

		// Step 2: a matching pricelist rule overrides the catalogue price.
		//
		// A rule that matches but cannot compute — a FORMULA whose cost or base list is missing —
		// leaves the catalogue price standing rather than refusing the line. See rulePrice.
		if item, found := bestRule(input.PricelistItems, line); found {
			if price, computed := rulePrice(item, line, scale); computed {
				result.BaseUnitPrice = price
				result.EffectiveUnitPrice = price
				result.PricingSource = "pricelist"
				result.PricelistItemId = item.Id
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

// expandComponents allocates the combo price across its components (BR §18, D-04).
//
// The allocation is an OUTPUT — for VAT, partial return and reporting — never an input to the combo
// price, which is independent by BR §15. Its sum equals the line's gross amount exactly.
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
	// The allocated amounts already sum to the combo price, so they cannot serve as the reference.
	// A combo line's reference is its gross amount, which makes the combo adjustment zero unless
	// the caller supplied standalone prices — see the test for the worked example.
	return line.GrossAmount
}

// applyPrograms runs steps 4–5.
func applyPrograms(
	input Input, lines []LineResult, adjustments []Adjustment, sequence int32, scale int32,
) ([]LineResult, []Adjustment, int32) {
	// The pre-application state every program is evaluated against. See Calculate's doc comment for
	// why this is captured once rather than recomputed as programs apply.
	baseNet := netTotalOf(lines)

	for _, program := range input.Programs {
		rewards := make([]RewardInput, len(program.Rewards))
		copy(rewards, program.Rewards)
		sort.SliceStable(rewards, func(i, j int) bool {
			return rewards[i].Sequence < rewards[j].Sequence
		})

		for _, reward := range rewards {
			sequence++
			lines, adjustments = applyReward(
				lines, adjustments, program, reward, baseNet, sequence, scale)
		}
	}
	return lines, adjustments, sequence
}

// applyReward applies one reward and records what it did.
func applyReward(
	lines []LineResult, adjustments []Adjustment,
	program AppliedProgram, reward RewardInput,
	baseNet decimal.Decimal, sequence int32, scale int32,
) ([]LineResult, []Adjustment) {
	description := program.ProgramName
	if program.VoucherCode != "" {
		description = program.ProgramName + " (" + program.VoucherCode + ")"
	}
	adjustmentType := AdjustmentPercentageDiscount
	if program.VoucherCode != "" {
		adjustmentType = AdjustmentVoucher
	}

	switch reward.Type {
	case "percentage_discount":
		targets := targetIndices(lines, reward)
		base := netOfIndices(lines, targets)
		amount := base.Mul(reward.Value).Div(decimal.NewFromInt(100)).Round(scale).Neg()
		lines = applyDiscountToLines(lines, targets, amount, scale)
		adjustments = append(adjustments, Adjustment{
			Sequence: sequence, Type: adjustmentType,
			SourceType: "sales_promotion_program", SourceId: program.ProgramId,
			Description: description, BaseAmount: base, Amount: amount,
		})

	case "fixed_amount_discount":
		targets := targetIndices(lines, reward)
		base := netOfIndices(lines, targets)
		// Never discount more than the target lines are worth: a fixed amount larger than the
		// basket would otherwise produce a negative total and pay the customer to shop.
		amount := reward.Value.Round(scale)
		if amount.GreaterThan(base) {
			amount = base
		}
		amount = amount.Neg()
		lines = applyDiscountToLines(lines, targets, amount, scale)
		adjustments = append(adjustments, Adjustment{
			Sequence: sequence, Type: chooseFixedType(adjustmentType),
			SourceType: "sales_promotion_program", SourceId: program.ProgramId,
			Description: description, BaseAmount: base, Amount: amount,
		})

	case "fixed_product_price":
		// BR §20's conditional bundle pricing: not a combo, but a program whose reward sets a
		// per-unit price on particular lines.
		targets := targetIndices(lines, reward)
		base := netOfIndices(lines, targets)
		for _, index := range targets {
			newGross := reward.Value.Mul(lines[index].Quantity).Round(scale)
			lines[index].EffectiveUnitPrice = reward.Value
			lines[index].DiscountAmount = lines[index].DiscountAmount.Add(
				lines[index].NetAmount.Sub(newGross))
			lines[index].NetAmount = newGross
			lines[index].PricingSource = "pricelist"
		}
		after := netOfIndices(lines, targets)
		adjustments = append(adjustments, Adjustment{
			Sequence: sequence, Type: AdjustmentConditionalPrice,
			SourceType: "sales_promotion_program", SourceId: program.ProgramId,
			Description: description, BaseAmount: base, Amount: after.Sub(base),
		})

	case "free_quantity":
		// D-11: a free item is a REAL line at zero price, not an adjustment. Inventory must
		// physically fulfil it, and its VAT treatment is a line-level question.
		lines = append(lines, LineResult{
			Key:                      reward.RewardId,
			LineNumber:               nextLineNumber(lines),
			ProductVariantId:         reward.FreeProductVariantId,
			UomId:                    reward.FreeUomId,
			Quantity:                 reward.Value,
			ProductCode:              reward.FreeProductCode,
			ProductName:              reward.FreeProductName,
			LineType:                 "promotion_reward",
			PricingSource:            "promotion_reward",
			SourcePromotionProgramId: program.ProgramId,
			RequiresFulfillment:      true,
			BaseUnitPrice:            decimal.Zero,
			EffectiveUnitPrice:       decimal.Zero,
		})
	}
	return lines, adjustments
}

func chooseFixedType(adjustmentType string) string {
	if adjustmentType == AdjustmentVoucher {
		return AdjustmentVoucher
	}
	return AdjustmentFixedDiscount
}

// targetIndices resolves which lines a reward applies to. An order-scoped reward, or one naming no
// lines, applies to every line that carries a price.
func targetIndices(lines []LineResult, reward RewardInput) []int {
	wanted := make(map[string]bool, len(reward.TargetLineKeys))
	for _, key := range reward.TargetLineKeys {
		wanted[key] = true
	}

	indices := make([]int, 0, len(lines))
	for index, line := range lines {
		if len(wanted) > 0 && !wanted[line.Key] {
			continue
		}
		// A giveaway line is never itself discounted: it is already free, and allocating a share of
		// a discount to it would take that share away from a line the customer is actually paying
		// for.
		if line.LineType == "promotion_reward" {
			continue
		}
		indices = append(indices, index)
	}
	return indices
}

// applyDiscountToLines spreads a discount across the target lines proportionally to their CURRENT
// net amounts (D-05).
//
// Post-discount net rather than gross, deliberately: it keeps Σ line net equal to the order net at
// every step, and stops a line already discounted to near zero from absorbing more discount than it
// is worth.
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

// allocateOrderAdjustments is step 6. The discounts have already been pushed down to the lines as
// they were applied, so this step reconciles rather than re-allocates: it ensures each line's net is
// its gross minus its accumulated discount, and never negative.
func allocateOrderAdjustments(
	lines []LineResult, _ []Adjustment, scale int32,
) []LineResult {
	for index := range lines {
		net := lines[index].GrossAmount.Sub(lines[index].DiscountAmount).Round(scale)
		if net.IsNegative() {
			// A discount cannot take a line below zero. The cap is applied here rather than at each
			// reward so that a stack of rewards is capped once, on their combined effect.
			lines[index].DiscountAmount = lines[index].GrossAmount
			net = decimal.Zero
		}
		lines[index].NetAmount = net
	}
	return lines
}

// applyTax is step 7. It copies in what Accounting decided; it does not compute tax.
//
// This function used to extract tax itself, as net × rate / (1 + rate). That arithmetic now lives in
// accounting/domain/services/tax/calculation.go, which does exactly this for a tax-inclusive
// percentage tax and also handles the cases this engine never did: compound taxes whose base
// includes the tax before them, division-type taxes, fixed per-unit taxes, and document-scoped
// rounding. Two implementations of the same law would eventually disagree, and the one embedded in
// a pricing engine is the one that would be wrong.
//
// D-02 is settled by that delegation rather than decided here. Whether the price includes tax is
// Accounting's price_inclusion_mode plus the request's price_mode; Sales no longer holds an opinion.
//
// A line with no entry in the map is taxed at zero — see the note on Context.TaxByLineKey.
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

		// Accounting taxed the components individually, so its split is authoritative: it is the one
		// that reconciles against the rounding policy actually applied.
		if len(tax.ComponentAmounts) > 0 {
			lines[index].Components = copyComponentTax(lines[index].Components, tax.ComponentAmounts)
			continue
		}

		// Otherwise the line was taxed as a whole and the components share it in proportion to their
		// allocated net. The allocator guarantees the shares sum to the line's tax exactly, so a VAT
		// invoice itemising a combo still adds up to what was charged.
		lines[index].Components = allocateComponentTax(
			lines[index].Components, lines[index].TaxAmount, scale)
	}
	return lines
}

// copyComponentTax takes Accounting's per-component figures.
//
// A component Accounting did not report gets zero rather than a guessed share: inventing a number to
// fill a gap is how a total stops reconciling with the tax that was actually charged.
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

// totalise sums the lines into the order totals.
//
// FinalAmount is the line's net: with tax-inclusive pricing the tax is already inside it, so adding
// tax again would charge it twice.
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

func netOfIndices(lines []LineResult, indices []int) decimal.Decimal {
	total := decimal.Zero
	for _, index := range indices {
		total = total.Add(lines[index].NetAmount)
	}
	return total
}

func nextLineNumber(lines []LineResult) int32 {
	highest := int32(0)
	for _, line := range lines {
		if line.LineNumber > highest {
			highest = line.LineNumber
		}
	}
	return highest + 1
}
