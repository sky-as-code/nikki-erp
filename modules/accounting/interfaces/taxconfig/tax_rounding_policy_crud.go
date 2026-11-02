package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRoundingPolicyRepository interface {
	dynamicresource.CrudRepository
}

type TaxRoundingPolicyDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxRoundingPolicyApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxRoundingPolicyCommand = dynamicresource.CreateCommand
	UpdateTaxRoundingPolicyCommand = dynamicresource.UpdateCommand
	DeleteTaxRoundingPolicyCommand = dynamicresource.DeleteCommand
	SearchTaxRoundingPolicyQuery   = dynamicresource.SearchQuery
)
