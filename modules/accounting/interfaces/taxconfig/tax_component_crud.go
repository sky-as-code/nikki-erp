package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxComponentRepository interface {
	composable.CrudRepository
}

type TaxComponentDomainService interface {
	composable.CrudDomainService
}

type TaxComponentApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxComponentCommand = composable.CreateCommand
	UpdateTaxComponentCommand = composable.UpdateCommand
	DeleteTaxComponentCommand = composable.DeleteCommand
	SearchTaxComponentQuery   = composable.SearchQuery
)
