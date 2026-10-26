// Package catalog declares the layers of the Sales catalogue resources: the fulfillment methods,
// channels and points a sale happens through, and the pricelists and combos it is priced by.
//
// Each resource publishes three interfaces -- repository, domain service, application service --
// so the onion's layers are injected by type rather than by concrete struct, and command and
// result aliases so a signature reads as the resource's own rather than as a bare field map.
package catalog

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SalesComboRepository reads and writes combos, the bundles sold as a single order line.
type SalesComboRepository interface {
	composable.CrudRepository
}

// SalesComboDomainService is the CRUD of the resource plus its own rules.
type SalesComboDomainService interface {
	composable.CrudDomainService
}

// SalesComboApplicationService is the authorized surface the REST handler serves.
type SalesComboApplicationService interface {
	composable.CrudApplicationService
}

// SalesComboComponentRepository reads and writes the products a combo is made of.
type SalesComboComponentRepository interface {
	composable.CrudRepository
}

// SalesComboComponentDomainService is the CRUD of the resource plus its own rules.
type SalesComboComponentDomainService interface {
	composable.CrudDomainService
}

// SalesComboComponentApplicationService is the authorized surface the REST handler serves.
type SalesComboComponentApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSalesComboCommand      = composable.CreateCommand
	UpdateSalesComboCommand      = composable.UpdateCommand
	DeleteSalesComboCommand      = composable.DeleteCommand
	SetSalesComboArchivedCommand = composable.SetArchivedCommand
	GetSalesComboByIdQuery       = composable.GetByIdQuery
	SearchSalesCombosQuery       = composable.SearchQuery
	SalesComboExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesComboResult      = composable.CreateResult
	UpdateSalesComboResult      = composable.MutateResult
	DeleteSalesComboResult      = composable.MutateResult
	SetSalesComboArchivedResult = composable.MutateResult
	GetSalesComboByIdResult     = composable.GetOneResult
	SearchSalesCombosResult     = composable.SearchResult
	SalesComboExistsResult      = composable.ExistsResult
)

type (
	CreateSalesComboComponentCommand      = composable.CreateCommand
	UpdateSalesComboComponentCommand      = composable.UpdateCommand
	DeleteSalesComboComponentCommand      = composable.DeleteCommand
	SetSalesComboComponentArchivedCommand = composable.SetArchivedCommand
	GetSalesComboComponentByIdQuery       = composable.GetByIdQuery
	SearchSalesComboComponentsQuery       = composable.SearchQuery
	SalesComboComponentExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesComboComponentResult      = composable.CreateResult
	UpdateSalesComboComponentResult      = composable.MutateResult
	DeleteSalesComboComponentResult      = composable.MutateResult
	SetSalesComboComponentArchivedResult = composable.MutateResult
	GetSalesComboComponentByIdResult     = composable.GetOneResult
	SearchSalesComboComponentsResult     = composable.SearchResult
	SalesComboComponentExistsResult      = composable.ExistsResult
)
