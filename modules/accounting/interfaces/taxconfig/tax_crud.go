package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxRepository interface {
	dynamicresource.CrudRepository
}

type TaxDomainService interface {
	dynamicresource.CrudDomainService
}

// TaxApplicationService is the authorized CRUD of the tax resource. The calculation actions
// served on it (calculate, simulate, reverse_full, reverse_partial) are TaxCalculationActions.
type TaxApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxCommand = dynamicresource.CreateCommand
	UpdateTaxCommand = dynamicresource.UpdateCommand
	DeleteTaxCommand = dynamicresource.DeleteCommand
	SearchTaxQuery   = dynamicresource.SearchQuery
)
