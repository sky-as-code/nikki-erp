package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxMappingLineRepository interface {
	dynamicresource.CrudRepository
}

type TaxMappingLineDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxMappingLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxMappingLineCommand = dynamicresource.CreateCommand
	UpdateTaxMappingLineCommand = dynamicresource.UpdateCommand
	DeleteTaxMappingLineCommand = dynamicresource.DeleteCommand
	SearchTaxMappingLineQuery   = dynamicresource.SearchQuery
)
