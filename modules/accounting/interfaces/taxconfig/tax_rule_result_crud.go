package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRuleResultRepository interface {
	dynamicresource.CrudRepository
}

type TaxRuleResultDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxRuleResultApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxRuleResultCommand = dynamicresource.CreateCommand
	UpdateTaxRuleResultCommand = dynamicresource.UpdateCommand
	DeleteTaxRuleResultCommand = dynamicresource.DeleteCommand
	SearchTaxRuleResultQuery   = dynamicresource.SearchQuery
)
