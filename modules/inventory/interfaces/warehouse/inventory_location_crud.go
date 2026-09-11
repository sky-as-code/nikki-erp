package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// InventoryLocationRepository reads and writes inventory_location rows.
type InventoryLocationRepository interface {
	composable.CrudRepository
}

// InventoryLocationDomainService is the CRUD of the resource plus its own rules.
type InventoryLocationDomainService interface {
	composable.CrudDomainService
}

// InventoryLocationApplicationService is the authorized surface the REST handler serves.
type InventoryLocationApplicationService interface {
	composable.CrudApplicationService
	InventoryLocationActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateInventoryLocationCommand      = composable.CreateCommand
	UpdateInventoryLocationCommand      = composable.UpdateCommand
	DeleteInventoryLocationCommand      = composable.DeleteCommand
	SetInventoryLocationArchivedCommand = composable.SetArchivedCommand
	GetInventoryLocationByIdQuery       = composable.GetByIdQuery
	SearchInventoryLocationsQuery       = composable.SearchQuery
	InventoryLocationExistsQuery        = composable.ExistsQuery
)

type (
	CreateInventoryLocationResult      = composable.CreateResult
	UpdateInventoryLocationResult      = composable.MutateResult
	DeleteInventoryLocationResult      = composable.MutateResult
	SetInventoryLocationArchivedResult = composable.MutateResult
	GetInventoryLocationByIdResult     = composable.GetOneResult
	SearchInventoryLocationsResult     = composable.SearchResult
	InventoryLocationExistsResult      = composable.ExistsResult
)
