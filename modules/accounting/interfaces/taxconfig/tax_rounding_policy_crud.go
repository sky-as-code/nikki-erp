package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRoundingPolicyRepository interface {
	composable.CrudRepository
}

type TaxRoundingPolicyDomainService interface {
	composable.CrudDomainService
}

type TaxRoundingPolicyApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxRoundingPolicyCommand = composable.CreateCommand
	UpdateTaxRoundingPolicyCommand = composable.UpdateCommand
	DeleteTaxRoundingPolicyCommand = composable.DeleteCommand
	SearchTaxRoundingPolicyQuery   = composable.SearchQuery
)
