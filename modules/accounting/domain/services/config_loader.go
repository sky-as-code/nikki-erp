package services

import (
	"sort"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
)

// Bounds on how much configuration one calculation will read: guards against a runaway query, not
// business limits. The loader reports truncation rather than proceeding on a partial rule set,
// because a rule that was not read is a rule that did not fire.
const (
	maxPublishedRules    = 500
	maxConditionsPerRule = 50
	maxResultsPerRule    = 50
	maxLinesPerMapping   = 200
)

// LoadEffectiveRules reads every published rule in force on a date, with its conditions and results.
// Rules come back unsorted; the determination engine sorts them by priority. The effective-period
// filter runs in memory because a nullable upper bound is awkward in the search graph.
func LoadEffectiveRules(
	ctx corectx.Context, repos *TaxRepos, taxDate string,
) ([]taxsvc.Rule, error) {
	rows, err := models.FindPublishedRules(ctx, repos.Rule, maxPublishedRules)
	if err != nil {
		return nil, err
	}

	rules := make([]taxsvc.Rule, 0, len(rows))
	for _, row := range rows {
		stored := models.NewTaxRuleFrom(row)
		if !models.PeriodContains(stored.GetEffectiveFrom(), stored.GetEffectiveTo(), taxDate) {
			continue
		}

		ruleId := idString(stored.GetId())
		conditions, err := loadConditions(ctx, repos.RuleCondition, ruleId)
		if err != nil {
			return nil, err
		}
		results, err := loadResults(ctx, repos.RuleResult, ruleId)
		if err != nil {
			return nil, err
		}

		rules = append(rules, taxsvc.Rule{
			Id:             ruleId,
			Code:           derefString(stored.GetCode()),
			Priority:       derefInt32(stored.GetPriority()),
			StopProcessing: derefBool(stored.GetStopProcessing()),
			Conditions:     conditions,
			Results:        results,
		})
	}
	return rules, nil
}

func loadConditions(
	ctx corectx.Context, repo models.TaxSearcher, ruleId string,
) ([]taxsvc.RuleCondition, error) {
	rows, err := models.FindConditionsOfRule(ctx, repo, ruleId, maxConditionsPerRule)
	if err != nil {
		return nil, err
	}

	conditions := make([]taxsvc.RuleCondition, 0, len(rows))
	for _, row := range rows {
		stored := models.NewTaxRuleConditionFrom(row)
		operator := models.ConditionOperator(derefString(stored.GetOperator()))
		scalar, list := conditionValue(stored.GetValue())

		condition := taxsvc.RuleCondition{
			FieldKey:     derefString(stored.GetFieldKey()),
			Operator:     operator,
			Value:        scalar,
			CurrencyCode: derefString(stored.GetValueCurrencyCode()),
		}
		// The list operators read the same stored column as the scalar ones. A JSON array arrives
		// already split; a plain string is split on commas, so "a,b" means two values.
		if operator == models.OperatorIn || operator == models.OperatorNotIn {
			if list != nil {
				condition.Values = list
			} else {
				condition.Values = splitList(scalar)
			}
		}
		conditions = append(conditions, condition)
	}
	return conditions, nil
}

func loadResults(
	ctx corectx.Context, repo models.TaxSearcher, ruleId string,
) ([]taxsvc.RuleResult, error) {
	rows, err := models.FindResultsOfRule(ctx, repo, ruleId, maxResultsPerRule)
	if err != nil {
		return nil, err
	}

	results := make([]taxsvc.RuleResult, 0, len(rows))
	for _, row := range rows {
		stored := models.NewTaxRuleResultFrom(row)
		results = append(results, taxsvc.RuleResult{
			Action:    models.RuleResultAction(derefString(stored.GetAction())),
			TaxId:     idString(stored.GetTaxId()),
			MappingId: idString(stored.GetTaxMappingId()),
			Treatment: models.TaxTreatment(derefString(stored.GetTaxTreatment())),
			Sequence:  derefInt32(stored.GetSequence()),
		})
	}
	return results, nil
}

// LoadMapping reads one mapping and its lines, returning nil when it does not exist. The caller
// must treat that as an unresolved determination, not as "no substitution": quietly applying no
// mapping would charge the un-substituted tax.
func LoadMapping(
	ctx corectx.Context, repos *TaxRepos, mappingId string,
) (*taxsvc.Mapping, error) {
	stored, err := models.FindMappingById(ctx, repos.Mapping, mappingId)
	if err != nil || stored == nil {
		return nil, err
	}

	rows, err := models.FindLinesOfMapping(ctx, repos.MappingLine, mappingId, maxLinesPerMapping)
	if err != nil {
		return nil, err
	}

	lines := make([]taxsvc.MappingLine, 0, len(rows))
	for _, row := range rows {
		line := models.NewTaxMappingLineFrom(row)
		lines = append(lines, taxsvc.MappingLine{
			SourceTaxId: idString(line.GetSourceTaxId()),
			TargetTaxId: idString(line.GetTargetTaxId()),
			Sequence:    derefInt32(line.GetSequence()),
		})
	}
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].Sequence < lines[j].Sequence })

	return &taxsvc.Mapping{
		Id:        mappingId,
		VersionNo: derefInt32(stored.GetVersionNo()),
		Lines:     lines,
	}, nil
}

// LoadMappingsForRules loads every mapping the given rules can reach. Loading up front keeps
// determination a pure function of its input, performing no I/O of its own.
func LoadMappingsForRules(
	ctx corectx.Context, repos *TaxRepos, rules []taxsvc.Rule,
) (map[string]taxsvc.Mapping, error) {
	mappings := map[string]taxsvc.Mapping{}
	for _, rule := range rules {
		for _, result := range rule.Results {
			if result.MappingId == "" {
				continue
			}
			if _, seen := mappings[result.MappingId]; seen {
				continue
			}
			mapping, err := LoadMapping(ctx, repos, result.MappingId)
			if err != nil {
				return nil, err
			}
			if mapping != nil {
				mappings[result.MappingId] = *mapping
			}
		}
	}
	return mappings, nil
}

// splitList turns a stored comma-separated list into its values, dropping empties and surrounding
// whitespace, so "a, b," is two values.
func splitList(value string) []string {
	if value == "" {
		return nil
	}

	values := make([]string, 0, 4)
	start := 0
	for index := 0; index <= len(value); index++ {
		if index < len(value) && value[index] != ',' {
			continue
		}
		if trimmed := trimSpace(value[start:index]); trimmed != "" {
			values = append(values, trimmed)
		}
		start = index + 1
	}
	return values
}

func trimSpace(value string) string {
	start := 0
	end := len(value)
	for start < end && isSpace(value[start]) {
		start++
	}
	for end > start && isSpace(value[end-1]) {
		end--
	}
	return value[start:end]
}

func isSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}
