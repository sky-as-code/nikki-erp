package tax

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

func addTaxRule(id string, priority int32, taxId string, conditions ...RuleCondition) Rule {
	return Rule{
		Id:         id,
		Code:       id,
		Priority:   priority,
		Conditions: conditions,
		Results: []RuleResult{
			{Action: models.ActionAddTax, TaxId: taxId, Sequence: 1},
		},
	}
}

func eq(fieldKey string, value string) RuleCondition {
	return RuleCondition{FieldKey: fieldKey, Operator: models.OperatorEq, Value: value}
}

// A candidate tax survives when no rule changes it.
func TestCandidateSurvivesWithoutRules(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10"},
		Context:         map[string]string{},
	})

	assert.Equal(t, models.DeterminationResolved, outcome.Status)
	assert.Equal(t, []string{"VAT10"}, outcome.TaxIds)
}

// Nothing determined is unresolved, never a silent 0%. A tax that quietly resolves to zero produces
// an invoice that looks correct, under-collects, and is indistinguishable downstream from a genuine
// zero-rating.
func TestEmptyCandidateWithNoRuleIsUnresolved(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{},
		Context:         map[string]string{},
	})

	assert.Equal(t, models.DeterminationUnresolved, outcome.Status)
	assert.Equal(t, models.ErrCodeNoApplicableTax, outcome.ErrorCode)
	assert.Empty(t, outcome.TaxIds)
}

// An explicit no_tax_applicable is a conclusion, not a failure.
func TestExplicitNoTaxApplicableIsNotUnresolved(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{},
		Context:         map[string]string{models.CtxProductTaxClassification: "EXPORT_SERVICE"},
		Rules: []Rule{{
			Id:       "R1",
			Priority: 10,
			Conditions: []RuleCondition{
				eq(models.CtxProductTaxClassification, "EXPORT_SERVICE"),
			},
			Results: []RuleResult{{
				Action:    models.ActionNoTaxApplicable,
				Treatment: models.TaxTreatmentOutOfScope,
				Sequence:  1,
			}},
		}},
	})

	assert.Equal(t, models.DeterminationNoTaxApplicable, outcome.Status)
	assert.Equal(t, models.TaxTreatmentOutOfScope, outcome.Treatment)
	assert.NotEqual(t, models.DeterminationUnresolved, outcome.Status,
		"unresolved and no_tax_applicable are different answers")
}

// Evaluation order is priority then id, with no specificity heuristic: the rule the author numbered
// first wins even though the other has more conditions.
func TestRulesEvaluateByPriorityNotSpecificity(t *testing.T) {
	specific := addTaxRule("R_SPECIFIC", 50, "VAT_SPECIFIC",
		eq(models.CtxProductTaxClassification, "STANDARD"),
		eq(models.CtxCurrencyCode, "VND"),
		eq(models.CtxOperationType, "sale"))
	general := addTaxRule("R_GENERAL", 10, "VAT_GENERAL")
	general.StopProcessing = true

	outcome := Determine(DeterminationInput{
		Context: map[string]string{
			models.CtxProductTaxClassification: "STANDARD",
			models.CtxCurrencyCode:             "VND",
			models.CtxOperationType:            "sale",
		},
		Rules: []Rule{specific, general},
	})

	assert.Equal(t, []string{"VAT_GENERAL"}, outcome.TaxIds)
	assert.Equal(t, []string{"R_GENERAL"}, outcome.AppliedRuleIds)
}

// stop_processing ends evaluation, so lower-priority rules never contribute.
func TestStopProcessingHaltsEvaluation(t *testing.T) {
	first := addTaxRule("R1", 10, "VAT_A")
	first.StopProcessing = true

	outcome := Determine(DeterminationInput{
		Rules:   []Rule{first, addTaxRule("R2", 20, "VAT_B")},
		Context: map[string]string{},
	})

	assert.Equal(t, []string{"VAT_A"}, outcome.TaxIds)
}

// The Vietnam VAT reduction: remove the standard rate, add the reduced one. Result sequence is
// load-bearing; reversing it would remove the tax that was just added.
func TestVietnamVatReductionSubstitutesRate(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VN_VAT_10"},
		Context: map[string]string{
			models.CtxProductTaxClassification: "VAT_REDUCTION_ELIGIBLE",
		},
		Rules: []Rule{{
			Id:       "VN-VAT-REDUCTION-2025",
			Priority: 20,
			Conditions: []RuleCondition{
				eq(models.CtxProductTaxClassification, "VAT_REDUCTION_ELIGIBLE"),
			},
			Results: []RuleResult{
				{Action: models.ActionRemoveTax, TaxId: "VN_VAT_10", Sequence: 1},
				{Action: models.ActionAddTax, TaxId: "VN_VAT_8", Sequence: 2},
			},
		}},
	})

	assert.Equal(t, models.DeterminationResolved, outcome.Status)
	assert.Equal(t, []string{"VN_VAT_8"}, outcome.TaxIds)
}

// A mapping substitutes a tax, and only when a rule result asks for it.
func TestMappingSubstitutesTax(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10"},
		Context:         map[string]string{models.CtxPartyTaxClassification: "EXPORT"},
		Rules: []Rule{{
			Id:         "R_EXPORT",
			Priority:   10,
			Conditions: []RuleCondition{eq(models.CtxPartyTaxClassification, "EXPORT")},
			Results: []RuleResult{
				{Action: models.ActionApplyMapping, MappingId: "M_EXPORT", Sequence: 1},
			},
		}},
		MappingsById: map[string]Mapping{
			"M_EXPORT": {
				Id:        "M_EXPORT",
				VersionNo: 1,
				Lines:     []MappingLine{{SourceTaxId: "VAT10", TargetTaxId: "VAT0_EXPORT"}},
			},
		},
	})

	assert.Equal(t, []string{"VAT0_EXPORT"}, outcome.TaxIds)
	assert.Equal(t, "M_EXPORT", outcome.AppliedMappingId)
}

// Two different mappings in one determination scope are unresolvable rather than resolved by
// whichever rule ran first.
func TestTwoMappingsAreUnresolved(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10"},
		Context:         map[string]string{},
		Rules: []Rule{
			{Id: "R1", Priority: 10, Results: []RuleResult{
				{Action: models.ActionApplyMapping, MappingId: "M1", Sequence: 1}}},
			{Id: "R2", Priority: 20, Results: []RuleResult{
				{Action: models.ActionApplyMapping, MappingId: "M2", Sequence: 1}}},
		},
	})

	assert.Equal(t, models.DeterminationUnresolved, outcome.Status)
	assert.Equal(t, models.ErrCodeMultipleTaxMappings, outcome.ErrorCode)
}

// A mapping leaves taxes it does not name alone, rather than dropping them.
func TestMappingLeavesUnmappedTaxesAlone(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10", "EXCISE"},
		Context:         map[string]string{},
		Rules: []Rule{{Id: "R1", Priority: 10, Results: []RuleResult{
			{Action: models.ActionApplyMapping, MappingId: "M1", Sequence: 1}}}},
		MappingsById: map[string]Mapping{
			"M1": {Id: "M1", Lines: []MappingLine{{SourceTaxId: "VAT10", TargetTaxId: "VAT0"}}},
		},
	})

	assert.Equal(t, []string{"VAT0", "EXCISE"}, outcome.TaxIds)
}

// An override replaces the determined set and is applied last, so the user substitutes the finished
// determination.
func TestOverrideReplacesDeterminedSet(t *testing.T) {
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10"},
		Context:         map[string]string{},
		OverrideTaxIds:  []string{"VAT_EXEMPT"},
	})

	assert.Equal(t, []string{"VAT_EXEMPT"}, outcome.TaxIds)
}

// A money threshold in another currency does not match, and is not converted.
func TestMoneyConditionDoesNotConvertCurrency(t *testing.T) {
	rule := addTaxRule("R1", 10, "LUXURY", RuleCondition{
		FieldKey:     models.CtxCommercialBaseAmount,
		Operator:     models.OperatorGte,
		Value:        "1000000",
		CurrencyCode: "VND",
	})

	// The threshold is VND and the request USD; converting would need a rate nobody agreed on, so
	// the rule does not fire.
	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"VAT10"},
		CurrencyCode:    "USD",
		Context:         map[string]string{models.CtxCommercialBaseAmount: "5000"},
		Rules:           []Rule{rule},
	})

	assert.Equal(t, []string{"VAT10"}, outcome.TaxIds)
	assert.Empty(t, outcome.AppliedRuleIds)
}

// Numeric comparison must be numeric: a string ordering would rank 9 above 10.
func TestOrderingComparesNumerically(t *testing.T) {
	rule := addTaxRule("R1", 10, "LUXURY", RuleCondition{
		FieldKey: models.CtxCommercialBaseAmount,
		Operator: models.OperatorGte,
		Value:    "9",
	})

	outcome := Determine(DeterminationInput{
		Context: map[string]string{models.CtxCommercialBaseAmount: "10"},
		Rules:   []Rule{rule},
	})

	assert.Equal(t, []string{"LUXURY"}, outcome.TaxIds, "10 >= 9 numerically")
}

// Conditions within a rule are ANDed: one false condition stops the rule matching.
func TestConditionsAreAnded(t *testing.T) {
	rule := addTaxRule("R1", 10, "VAT",
		eq(models.CtxCurrencyCode, "VND"),
		eq(models.CtxOperationType, "purchase"))

	outcome := Determine(DeterminationInput{
		CandidateTaxIds: []string{"BASE"},
		Context: map[string]string{
			models.CtxCurrencyCode:  "VND",
			models.CtxOperationType: "sale",
		},
		Rules: []Rule{rule},
	})

	assert.Empty(t, outcome.AppliedRuleIds)
	assert.Equal(t, []string{"BASE"}, outcome.TaxIds)
}

// Determination is deterministic: evaluation sorts the rules itself, so feeding them in a different
// order must not change the outcome.
func TestDeterminationIsOrderIndependent(t *testing.T) {
	ruleA := addTaxRule("R_A", 20, "VAT_A")
	ruleB := addTaxRule("R_B", 10, "VAT_B")

	forward := Determine(DeterminationInput{Rules: []Rule{ruleA, ruleB}, Context: map[string]string{}})
	reverse := Determine(DeterminationInput{Rules: []Rule{ruleB, ruleA}, Context: map[string]string{}})

	assert.Equal(t, forward.TaxIds, reverse.TaxIds)
	assert.Equal(t, forward.AppliedRuleIds, reverse.AppliedRuleIds)
	assert.Equal(t, []string{"R_B", "R_A"}, forward.AppliedRuleIds, "sorted by priority")
}
