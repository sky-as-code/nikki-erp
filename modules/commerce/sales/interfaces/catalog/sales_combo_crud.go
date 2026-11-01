// Package catalog declares the layers of the Sales catalogue resources: the fulfillment methods,
// channels and points a sale happens through, and the pricelists and combos it is priced by.
//
// Each resource publishes three interfaces -- repository, domain service, application service --
// so the onion's layers are injected by type rather than by concrete struct, and command and
// result aliases so a signature reads as the resource's own rather than as a bare field map.
package catalog

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesComboRepository reads and writes combos, the bundles sold as a single order line.
type SalesComboRepository interface {
	dynamicresource.CrudRepository
}

// SalesComboDomainService is the CRUD of the resource plus its own rules.
type SalesComboDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesComboApplicationService is the authorized surface the REST handler serves.
type SalesComboApplicationService interface {
	dynamicresource.CrudApplicationService
}

// SalesComboComponentRepository reads and writes the products a combo is made of.
type SalesComboComponentRepository interface {
	dynamicresource.CrudRepository
}

// SalesComboComponentDomainService is the CRUD of the resource plus its own rules.
type SalesComboComponentDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesComboComponentApplicationService is the authorized surface the REST handler serves.
type SalesComboComponentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesComboCommand      = dynamicresource.CreateCommand
	UpdateSalesComboCommand      = dynamicresource.UpdateCommand
	DeleteSalesComboCommand      = dynamicresource.DeleteCommand
	SetSalesComboArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesComboByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesCombosQuery       = dynamicresource.SearchQuery
	SalesComboExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesComboResult      = dynamicresource.CreateResult
	UpdateSalesComboResult      = dynamicresource.MutateResult
	DeleteSalesComboResult      = dynamicresource.MutateResult
	SetSalesComboArchivedResult = dynamicresource.MutateResult
	GetSalesComboByIdResult     = dynamicresource.GetOneResult
	SearchSalesCombosResult     = dynamicresource.SearchResult
	SalesComboExistsResult      = dynamicresource.ExistsResult
)

type (
	CreateSalesComboComponentCommand      = dynamicresource.CreateCommand
	UpdateSalesComboComponentCommand      = dynamicresource.UpdateCommand
	DeleteSalesComboComponentCommand      = dynamicresource.DeleteCommand
	SetSalesComboComponentArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesComboComponentByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesComboComponentsQuery       = dynamicresource.SearchQuery
	SalesComboComponentExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesComboComponentResult      = dynamicresource.CreateResult
	UpdateSalesComboComponentResult      = dynamicresource.MutateResult
	DeleteSalesComboComponentResult      = dynamicresource.MutateResult
	SetSalesComboComponentArchivedResult = dynamicresource.MutateResult
	GetSalesComboComponentByIdResult     = dynamicresource.GetOneResult
	SearchSalesComboComponentsResult     = dynamicresource.SearchResult
	SalesComboComponentExistsResult      = dynamicresource.ExistsResult
)
