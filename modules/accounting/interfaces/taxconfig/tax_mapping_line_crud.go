package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxMappingLineRepository interface {
	composable.CrudRepository
}

type TaxMappingLineDomainService interface {
	composable.CrudDomainService
}

type TaxMappingLineApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxMappingLineCommand = composable.CreateCommand
	UpdateTaxMappingLineCommand = composable.UpdateCommand
	DeleteTaxMappingLineCommand = composable.DeleteCommand
	SearchTaxMappingLineQuery   = composable.SearchQuery
)
