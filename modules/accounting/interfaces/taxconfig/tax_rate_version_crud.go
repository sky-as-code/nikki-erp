package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRateVersionRepository interface {
	dynamicresource.CrudRepository
}

type TaxRateVersionDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxRateVersionApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxRateVersionCommand = dynamicresource.CreateCommand
	UpdateTaxRateVersionCommand = dynamicresource.UpdateCommand
	DeleteTaxRateVersionCommand = dynamicresource.DeleteCommand
	SearchTaxRateVersionQuery   = dynamicresource.SearchQuery
)
