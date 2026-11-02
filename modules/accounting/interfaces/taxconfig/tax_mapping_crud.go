package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxMappingRepository interface {
	dynamicresource.CrudRepository
}

type TaxMappingDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxMappingApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxMappingCommand = dynamicresource.CreateCommand
	UpdateTaxMappingCommand = dynamicresource.UpdateCommand
	DeleteTaxMappingCommand = dynamicresource.DeleteCommand
	SearchTaxMappingQuery   = dynamicresource.SearchQuery
)
