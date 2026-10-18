package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRuleRepository interface {
	composable.CrudRepository
}

type TaxRuleDomainService interface {
	composable.CrudDomainService
}

type TaxRuleApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxRuleCommand = composable.CreateCommand
	UpdateTaxRuleCommand = composable.UpdateCommand
	DeleteTaxRuleCommand = composable.DeleteCommand
	SearchTaxRuleQuery   = composable.SearchQuery
)
