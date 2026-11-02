package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRuleConditionRepository interface {
	dynamicresource.CrudRepository
}

type TaxRuleConditionDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxRuleConditionApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxRuleConditionCommand = dynamicresource.CreateCommand
	UpdateTaxRuleConditionCommand = dynamicresource.UpdateCommand
	DeleteTaxRuleConditionCommand = dynamicresource.DeleteCommand
	SearchTaxRuleConditionQuery   = dynamicresource.SearchQuery
)
