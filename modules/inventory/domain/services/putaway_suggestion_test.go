package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// A criterion a rule leaves empty matches anything, which is what makes a general rule general.
// One the rule names but the caller cannot answer does not match, because nothing shows it applies.
func TestCriterionMatches(t *testing.T) {
	assert.True(t, criterionMatches("", "anything"), "an unset criterion matches")
	assert.True(t, criterionMatches("", ""), "an unset criterion matches an unknown value too")
	assert.True(t, criterionMatches("prod-1", "prod-1"))
	assert.False(t, criterionMatches("prod-1", "prod-2"))
	assert.False(t, criterionMatches("prod-1", ""),
		"a rule asking about a product the caller did not name cannot be shown to apply")
}

func TestRuleMatchesContext(t *testing.T) {
	general := models.NewPutawayRuleFrom(dmodel.DynamicFields{})
	assert.True(t, ruleMatchesContext(*general, PutawayContext{ProductId: "prod-1"}),
		"a rule with no criteria applies to whatever arrives")

	productSpecific := models.NewPutawayRuleFrom(dmodel.DynamicFields{
		models.PutawayRuleFieldProductId: "prod-1",
	})
	assert.True(t, ruleMatchesContext(*productSpecific, PutawayContext{ProductId: "prod-1"}))
	assert.False(t, ruleMatchesContext(*productSpecific, PutawayContext{ProductId: "prod-2"}))
	assert.False(t, ruleMatchesContext(*productSpecific, PutawayContext{}))

	// Every named criterion has to match, not just one of them.
	twoCriteria := models.NewPutawayRuleFrom(dmodel.DynamicFields{
		models.PutawayRuleFieldProductId:     "prod-1",
		models.PutawayRuleFieldPackageTypeId: "pallet",
	})
	assert.True(t, ruleMatchesContext(*twoCriteria, PutawayContext{
		ProductId: "prod-1", PackageTypeId: "pallet",
	}))
	assert.False(t, ruleMatchesContext(*twoCriteria, PutawayContext{
		ProductId: "prod-1", PackageTypeId: "carton",
	}))
}

// Priority decides which of several matching rules wins, lowest first. The sort has to be stable
// so that two rules of equal priority resolve the same way every time rather than at random.
func TestPutawayRulePriorityOrdering(t *testing.T) {
	rules := []models.PutawayRule{
		*ruleWithPriority("late", 30),
		*ruleWithPriority("first", 1),
		*ruleWithPriority("middle", 10),
	}

	sortRulesByPriority(rules)

	assert.Equal(t, []string{"first", "middle", "late"}, ruleCodes(rules))
}

func TestPutawayRulePriorityOrderingIsStable(t *testing.T) {
	rules := []models.PutawayRule{
		*ruleWithPriority("a", 5),
		*ruleWithPriority("b", 5),
		*ruleWithPriority("c", 5),
	}

	sortRulesByPriority(rules)

	assert.Equal(t, []string{"a", "b", "c"}, ruleCodes(rules),
		"equal priorities keep the order they came in, so the answer is repeatable")
}

func ruleWithPriority(code string, priority int32) *models.PutawayRule {
	return models.NewPutawayRuleFrom(dmodel.DynamicFields{
		models.PutawayRuleFieldCode:     code,
		models.PutawayRuleFieldPriority: priority,
	})
}

func ruleCodes(rules []models.PutawayRule) []string {
	codes := make([]string, 0, len(rules))
	for _, rule := range rules {
		codes = append(codes, derefString(rule.GetCode()))
	}
	return codes
}
