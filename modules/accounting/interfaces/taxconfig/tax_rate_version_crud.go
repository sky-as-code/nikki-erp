package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRateVersionRepository interface {
	composable.CrudRepository
}

type TaxRateVersionDomainService interface {
	composable.CrudDomainService
}

type TaxRateVersionApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxRateVersionCommand = composable.CreateCommand
	UpdateTaxRateVersionCommand = composable.UpdateCommand
	DeleteTaxRateVersionCommand = composable.DeleteCommand
	SearchTaxRateVersionQuery   = composable.SearchQuery
)
