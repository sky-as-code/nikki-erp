package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxComponentRepository interface {
	dynamicresource.CrudRepository
}

type TaxComponentDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxComponentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxComponentCommand = dynamicresource.CreateCommand
	UpdateTaxComponentCommand = dynamicresource.UpdateCommand
	DeleteTaxComponentCommand = dynamicresource.DeleteCommand
	SearchTaxComponentQuery   = dynamicresource.SearchQuery
)
