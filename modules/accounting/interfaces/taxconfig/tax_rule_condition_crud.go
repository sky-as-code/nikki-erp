package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRuleConditionRepository interface {
	composable.CrudRepository
}

type TaxRuleConditionDomainService interface {
	composable.CrudDomainService
}

type TaxRuleConditionApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxRuleConditionCommand = composable.CreateCommand
	UpdateTaxRuleConditionCommand = composable.UpdateCommand
	DeleteTaxRuleConditionCommand = composable.DeleteCommand
	SearchTaxRuleConditionQuery   = composable.SearchQuery
)
