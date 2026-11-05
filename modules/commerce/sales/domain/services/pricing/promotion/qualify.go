package promotion

import (
	"sort"

	"github.com/shopspring/decimal"
)

// ruleEvaluation is one rule measured on the pre-promotion base.
type ruleEvaluation struct {
	rule      Rule
	lines     []Line
	quantity  decimal.Decimal
	amount    decimal.Decimal
	qualified bool
	reason    string
	sets      int64
}

// rulesOfPhase returns a phase's rules sorted by id, so "the first qualifying rule" is deterministic.
func rulesOfPhase(program Program, phase string) []Rule {
	rules := []Rule{}
	for _, rule := range program.Rules {
		if rule.Phase == phase {
			rules = append(rules, rule)
		}
	}
	sort.SliceStable(rules, func(i, j int) bool { return rules[i].Id < rules[j].Id })
	return rules
}

// measureRule applies BR-C04..C08. byQualificationUom switches the quantity to the rule's
// qualification UoM (BXGY); a line not convertible to it does not count (BR-C07).
func measureRule(input *Input, program Program, rule Rule, byQualificationUom bool) ruleEvaluation {
	lines, reason := selectLines(input, program, rule.Selector)
	if byQualificationUom {
		lines = convertibleLines(lines, rule.QualificationUomId)
		if len(lines) == 0 && reason == "" {
			reason = ReasonNoEligibleLine
		}
	}
	evaluation := ruleEvaluation{rule: rule, lines: lines, reason: reason}
	if len(lines) == 0 {
		return evaluation
	}
	evaluation.quantity = sumQuantity(lines, rule, byQualificationUom)
	evaluation.amount = sumAmount(lines, rule.AmountTaxBasis).Round(input.CurrencyScale)

	if evaluation.quantity.LessThan(rule.MinQuantity) || evaluation.amount.LessThan(rule.MinAmount) {
		evaluation.reason = ReasonConditionNotMet
		return evaluation
	}
	if byQualificationUom {
		// BR-C08: a BXGY rule without a positive X cannot count sets.
		if !rule.MinQuantity.IsPositive() {
			evaluation.reason = ReasonConditionNotMet
			return evaluation
		}
		evaluation.sets = evaluation.quantity.Div(rule.MinQuantity).Floor().IntPart()
	}
	evaluation.qualified = true
	return evaluation
}

func convertibleLines(lines []Line, uomId string) []Line {
	kept := []Line{}
	for _, line := range lines {
		if _, found := line.QualificationQuantities[uomId]; found {
			kept = append(kept, line)
		}
	}
	return kept
}

func sumQuantity(lines []Line, rule Rule, byQualificationUom bool) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		if byQualificationUom {
			total = total.Add(line.QualificationQuantities[rule.QualificationUomId])
			continue
		}
		total = total.Add(line.BaseQuantity)
	}
	return total
}

// sumAmount is ΣN_i, or Σ(N_i+T_i) for a tax-included threshold (BR-C06).
func sumAmount(lines []Line, basis string) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.UntaxedAmount)
		if basis == TaxBasisIncluded {
			total = total.Add(line.TaxAmount)
		}
	}
	return total
}

// qualifyFirst ORs the rules of a phase: the first qualifying rule by id decides (BR-C04). Nothing
// is summed across rules.
func qualifyFirst(input *Input, program Program, phase string) Qualification {
	rules := rulesOfPhase(program, phase)
	if len(rules) == 0 {
		return Qualification{ReasonCode: ReasonNoRule}
	}
	evaluations := make([]ruleEvaluation, 0, len(rules))
	for _, rule := range rules {
		evaluation := measureRule(input, program, rule, false)
		if evaluation.qualified {
			return qualificationOf(evaluation)
		}
		evaluations = append(evaluations, evaluation)
	}
	return Qualification{ReasonCode: strongestReason(evaluations)}
}

func qualificationOf(evaluation ruleEvaluation) Qualification {
	return Qualification{
		Qualified:        true,
		RuleId:           evaluation.rule.Id,
		EligibleLineKeys: lineKeys(evaluation.lines),
		EligibleQuantity: evaluation.quantity,
		EligibleAmount:   evaluation.amount,
		Sets:             evaluation.sets,
	}
}

// strongestReason explains a failure by the rule that came closest: a threshold miss beats a
// pricelist mismatch, which beats having no matching line at all.
func strongestReason(evaluations []ruleEvaluation) string {
	rank := map[string]int{ReasonConditionNotMet: 3, ReasonPricelistMismatch: 2, ReasonNoEligibleLine: 1}
	best := ReasonNoEligibleLine
	for _, evaluation := range evaluations {
		if rank[evaluation.reason] > rank[best] {
			best = evaluation.reason
		}
	}
	return best
}
