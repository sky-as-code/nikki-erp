package tax

import (
	"sort"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// Rule evaluation and tax-set transformation.
//
// Determination answers "which taxes apply", never "how much", and runs before calculation. It
// starts from the caller's candidate set, lets rules add and remove, optionally substitutes through
// a mapping, then hands the result to the calculator.

// RuleCondition is one predicate, resolved into plain values.
type RuleCondition struct {
	FieldKey string
	Operator models.ConditionOperator

	// Value is the comparison operand. For in and not_in it is the list; for is_null and
	// is_not_null it is unused.
	Value  string
	Values []string

	// CurrencyCode tags a money threshold. There is no FX conversion, so a condition in a currency
	// other than the request's does not match.
	CurrencyCode string
}

// RuleResult is one action a matching rule takes on the tax set.
type RuleResult struct {
	Action    models.RuleResultAction
	TaxId     string
	MappingId string
	Treatment models.TaxTreatment
	Sequence  int32
}

// Rule is one determination rule with its conditions and results.
type Rule struct {
	Id       string
	Code     string
	Priority int32

	// StopProcessing ends evaluation after this rule matches, so lower-priority rules never run.
	StopProcessing bool

	Conditions []RuleCondition
	Results    []RuleResult
}

// MappingLine substitutes one tax for another.
type MappingLine struct {
	SourceTaxId string
	TargetTaxId string
	Sequence    int32
}

// Mapping is a resolved tax mapping with its lines.
type Mapping struct {
	Id        string
	VersionNo int32
	Lines     []MappingLine
}

// DeterminationInput is everything determination needs to decide a tax set.
type DeterminationInput struct {
	// Context is the whitelisted facts a condition may test, keyed by the Ctx* constants.
	Context map[string]string

	// CurrencyCode is the request's currency, compared against a money condition's own.
	CurrencyCode string

	// CandidateTaxIds is the caller's proposal, typically a product's default tax. Rules may empty
	// it entirely.
	CandidateTaxIds []string

	// Rules are every rule in force at the request's tax date, in any order. Evaluation sorts them.
	Rules []Rule

	// MappingsById supplies a mapping when a rule result asks to apply one.
	MappingsById map[string]Mapping

	// OverrideTaxIds, when non-empty, replaces the determined set outright. It may only substitute
	// taxes, never a raw amount or an arbitrary rate.
	OverrideTaxIds []string
}

// DeterminationOutcome is the tax set plus the trace of how it was reached.
type DeterminationOutcome struct {
	Status models.DeterminationStatus

	// Treatment is set when Status is no_tax_applicable, and says which legal reason applies.
	Treatment models.TaxTreatment

	// ErrorCode lets a caller tell a missing rate from an ambiguous mapping without parsing prose.
	ErrorCode string

	TaxIds []string

	// AppliedRuleIds and AppliedMappingId are the audit trail the snapshot and simulator need.
	AppliedRuleIds   []string
	AppliedMappingId string
}

// Determine runs the tax-set transformation pipeline in a fixed order: candidates, then rules, then
// at most one mapping, then an authorized override. One pipeline rather than three engines, so they
// cannot disagree; mapping is not a determination engine of its own.
func Determine(input DeterminationInput) DeterminationOutcome {
	outcome := DeterminationOutcome{
		TaxIds:         append([]string{}, input.CandidateTaxIds...),
		AppliedRuleIds: []string{},
	}

	matched := evaluateRules(input)

	mappingId := ""
	noTaxTreatment := models.TaxTreatment("")
	sawNoTaxApplicable := false

	for _, rule := range matched {
		outcome.AppliedRuleIds = append(outcome.AppliedRuleIds, rule.Id)

		for _, result := range sortedResults(rule.Results) {
			switch result.Action {
			case models.ActionAddTax:
				outcome.TaxIds = addTax(outcome.TaxIds, result.TaxId)
			case models.ActionRemoveTax:
				outcome.TaxIds = removeTax(outcome.TaxIds, result.TaxId)
			case models.ActionApplyMapping:
				// Two mappings in scope is unresolvable, not a race won by whichever rule ran first:
				// they may substitute the same tax for different targets and guessing is silently
				// wrong.
				if mappingId != "" && mappingId != result.MappingId {
					return DeterminationOutcome{
						Status:         models.DeterminationUnresolved,
						ErrorCode:      models.ErrCodeMultipleTaxMappings,
						AppliedRuleIds: outcome.AppliedRuleIds,
					}
				}
				mappingId = result.MappingId
			case models.ActionNoTaxApplicable:
				sawNoTaxApplicable = true
				noTaxTreatment = result.Treatment
			}
		}
	}

	if sawNoTaxApplicable {
		return DeterminationOutcome{
			Status:         models.DeterminationNoTaxApplicable,
			Treatment:      noTaxTreatment,
			TaxIds:         []string{},
			AppliedRuleIds: outcome.AppliedRuleIds,
		}
	}

	if mappingId != "" {
		mapping, exists := input.MappingsById[mappingId]
		if !exists {
			return DeterminationOutcome{
				Status:         models.DeterminationUnresolved,
				ErrorCode:      models.ErrCodeNoApplicableTax,
				AppliedRuleIds: outcome.AppliedRuleIds,
			}
		}
		outcome.TaxIds = applyMapping(outcome.TaxIds, mapping)
		outcome.AppliedMappingId = mappingId
	}

	// The override runs last, so a user substitutes the finished determination rather than an
	// intermediate set the rules were still working on.
	if len(input.OverrideTaxIds) > 0 {
		outcome.TaxIds = append([]string{}, input.OverrideTaxIds...)
	}

	// An empty set is not "no tax due": nothing established the supply is outside tax, the
	// configuration just did not say. Defaulting it to zero would be invisible on the invoice.
	if len(outcome.TaxIds) == 0 {
		return DeterminationOutcome{
			Status:         models.DeterminationUnresolved,
			ErrorCode:      models.ErrCodeNoApplicableTax,
			AppliedRuleIds: outcome.AppliedRuleIds,
		}
	}

	outcome.Status = models.DeterminationResolved
	return outcome
}

// evaluateRules returns matching rules in evaluation order, honouring stop_processing. Order is
// priority ascending then id ascending, and nothing else: no specificity tie-break, since "more
// conditions means more specific" reorders rules whenever an author adds one. The id tie-break only
// makes the order total and reproducible.
func evaluateRules(input DeterminationInput) []Rule {
	candidates := append([]Rule{}, input.Rules...)
	sort.SliceStable(candidates, func(left, right int) bool {
		if candidates[left].Priority != candidates[right].Priority {
			return candidates[left].Priority < candidates[right].Priority
		}
		return candidates[left].Id < candidates[right].Id
	})

	matched := []Rule{}
	for _, rule := range candidates {
		if !ruleMatches(rule, input) {
			continue
		}
		matched = append(matched, rule)
		if rule.StopProcessing {
			break
		}
	}
	return matched
}

// ruleMatches reports whether every condition of a rule holds. Conditions are ANDed; OR is written
// as a second rule. A rule with no conditions matches everything, which is how a low-priority
// fallback is expressed.
func ruleMatches(rule Rule, input DeterminationInput) bool {
	for _, condition := range rule.Conditions {
		if !conditionHolds(condition, input) {
			return false
		}
	}
	return true
}

func conditionHolds(condition RuleCondition, input DeterminationInput) bool {
	actual, present := input.Context[condition.FieldKey]

	// A money threshold in another currency is not comparable and there is no FX contract, so it
	// fails to match rather than being converted at a rate nobody agreed.
	if condition.CurrencyCode != "" && condition.CurrencyCode != input.CurrencyCode {
		return false
	}

	switch condition.Operator {
	case models.OperatorIsNull:
		return !present || actual == ""
	case models.OperatorIsNotNull:
		return present && actual != ""
	case models.OperatorEq:
		return present && actual == condition.Value
	case models.OperatorNotEq:
		return !present || actual != condition.Value
	case models.OperatorIn:
		return present && containsValue(condition.Values, actual)
	case models.OperatorNotIn:
		return !present || !containsValue(condition.Values, actual)
	case models.OperatorGte:
		return present && compareValues(actual, condition.Value) >= 0
	case models.OperatorLte:
		return present && compareValues(actual, condition.Value) <= 0
	}
	return false
}

// compareValues orders two operands. Money and dates are the only orderable context fields: dates
// are ISO-8601 and sort lexicographically, money compares numerically when both sides parse, since
// string comparison would rank 9 above 10.
func compareValues(left string, right string) int {
	leftNumber, leftOk := parseDecimal(left)
	rightNumber, rightOk := parseDecimal(right)
	if leftOk && rightOk {
		return leftNumber.Cmp(rightNumber)
	}
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	}
	return 0
}

func containsValue(values []string, actual string) bool {
	for _, value := range values {
		if value == actual {
			return true
		}
	}
	return false
}

// sortedResults orders one rule's results by sequence. Load-bearing for substitution: removing VAT
// 10% then adding VAT 8% replaces one with the other, while the reverse order removes what was just
// added and leaves the line untaxed.
func sortedResults(results []RuleResult) []RuleResult {
	ordered := append([]RuleResult{}, results...)
	sort.SliceStable(ordered, func(left, right int) bool {
		return ordered[left].Sequence < ordered[right].Sequence
	})
	return ordered
}

func addTax(taxIds []string, taxId string) []string {
	if taxId == "" {
		return taxIds
	}
	for _, existing := range taxIds {
		if existing == taxId {
			return taxIds
		}
	}
	return append(taxIds, taxId)
}

func removeTax(taxIds []string, taxId string) []string {
	kept := make([]string, 0, len(taxIds))
	for _, existing := range taxIds {
		if existing != taxId {
			kept = append(kept, existing)
		}
	}
	return kept
}

// applyMapping substitutes each source tax for its target, leaving unmapped taxes alone.
func applyMapping(taxIds []string, mapping Mapping) []string {
	substitutions := map[string]string{}
	for _, line := range mapping.Lines {
		substitutions[line.SourceTaxId] = line.TargetTaxId
	}

	mapped := make([]string, 0, len(taxIds))
	for _, taxId := range taxIds {
		target, substituted := substitutions[taxId]
		if !substituted {
			mapped = append(mapped, taxId)
			continue
		}
		// Mapping a tax to nothing is legitimate: it expresses "not taxed by that tax here"
		// without inventing a zero-rate tax.
		if target == "" {
			continue
		}
		mapped = addTax(mapped, target)
	}
	return mapped
}

func parseDecimal(value string) (decimal.Decimal, bool) {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, false
	}
	return parsed, true
}
