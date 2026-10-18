package services

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
)

// TaxRepos is the set of repositories one calculation reads. Named fields make handing the wrong
// repository to a lookup a compile error rather than a runtime one: a repository answers only for
// the schema it was built for.
type TaxRepos struct {
	Tax               models.TaxSearcher
	DefinitionVersion models.TaxSearcher
	RateVersion       models.TaxSearcher
	Component         models.TaxSearcher
	RoundingPolicy    models.TaxSearcher
	Rule              models.TaxSearcher
	RuleCondition     models.TaxSearcher
	RuleResult        models.TaxSearcher
	Mapping           models.TaxSearcher
	MappingLine       models.TaxSearcher
}

// TaxReposParam receives the repositories the tax configuration onions publish.
type TaxReposParam struct {
	dig.In

	Tax               taxconfig.TaxRepository
	DefinitionVersion taxconfig.TaxDefinitionVersionRepository
	RateVersion       taxconfig.TaxRateVersionRepository
	Component         taxconfig.TaxComponentRepository
	RoundingPolicy    taxconfig.TaxRoundingPolicyRepository
	Rule              taxconfig.TaxRuleRepository
	RuleCondition     taxconfig.TaxRuleConditionRepository
	RuleResult        taxconfig.TaxRuleResultRepository
	Mapping           taxconfig.TaxMappingRepository
	MappingLine       taxconfig.TaxMappingLineRepository
}

// NewTaxRepos gathers the repositories once, at construction, instead of per request.
func NewTaxRepos(param TaxReposParam) *TaxRepos {
	return &TaxRepos{
		Tax:               param.Tax,
		DefinitionVersion: param.DefinitionVersion,
		RateVersion:       param.RateVersion,
		Component:         param.Component,
		RoundingPolicy:    param.RoundingPolicy,
		Rule:              param.Rule,
		RuleCondition:     param.RuleCondition,
		RuleResult:        param.RuleResult,
		Mapping:           param.Mapping,
		MappingLine:       param.MappingLine,
	}
}
