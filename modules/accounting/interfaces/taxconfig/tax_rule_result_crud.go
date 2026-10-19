package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRuleResultRepository interface {
	composable.CrudRepository
}

type TaxRuleResultDomainService interface {
	composable.CrudDomainService
}

type TaxRuleResultApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxRuleResultCommand = composable.CreateCommand
	UpdateTaxRuleResultCommand = composable.UpdateCommand
	DeleteTaxRuleResultCommand = composable.DeleteCommand
	SearchTaxRuleResultQuery   = composable.SearchQuery
)
