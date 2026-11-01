package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRuleRepository interface {
	dynamicresource.CrudRepository
}

type TaxRuleDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxRuleApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxRuleCommand = dynamicresource.CreateCommand
	UpdateTaxRuleCommand = dynamicresource.UpdateCommand
	DeleteTaxRuleCommand = dynamicresource.DeleteCommand
	SearchTaxRuleQuery   = dynamicresource.SearchQuery
)
