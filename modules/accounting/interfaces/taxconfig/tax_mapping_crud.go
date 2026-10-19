package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxMappingRepository interface {
	composable.CrudRepository
}

type TaxMappingDomainService interface {
	composable.CrudDomainService
}

type TaxMappingApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxMappingCommand = composable.CreateCommand
	UpdateTaxMappingCommand = composable.UpdateCommand
	DeleteTaxMappingCommand = composable.DeleteCommand
	SearchTaxMappingQuery   = composable.SearchQuery
)
