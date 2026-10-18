package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxRepository interface {
	composable.CrudRepository
}

type TaxDomainService interface {
	composable.CrudDomainService
}

// TaxApplicationService is the authorized CRUD of the tax resource. The calculation actions
// served on it (calculate, simulate, reverse_full, reverse_partial) are TaxCalculationActions.
type TaxApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxCommand = composable.CreateCommand
	UpdateTaxCommand = composable.UpdateCommand
	DeleteTaxCommand = composable.DeleteCommand
	SearchTaxQuery   = composable.SearchQuery
)
