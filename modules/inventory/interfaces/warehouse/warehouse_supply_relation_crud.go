package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// WarehouseSupplyRelationRepository reads and writes warehouse_supply_relation rows.
type WarehouseSupplyRelationRepository interface {
	composable.CrudRepository
}

// WarehouseSupplyRelationDomainService is the CRUD of the resource plus its own rules.
type WarehouseSupplyRelationDomainService interface {
	composable.CrudDomainService
}

// WarehouseSupplyRelationApplicationService is the authorized surface the REST handler serves.
type WarehouseSupplyRelationApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateWarehouseSupplyRelationCommand      = composable.CreateCommand
	UpdateWarehouseSupplyRelationCommand      = composable.UpdateCommand
	DeleteWarehouseSupplyRelationCommand      = composable.DeleteCommand
	SetWarehouseSupplyRelationArchivedCommand = composable.SetArchivedCommand
	GetWarehouseSupplyRelationByIdQuery       = composable.GetByIdQuery
	SearchWarehouseSupplyRelationsQuery       = composable.SearchQuery
	WarehouseSupplyRelationExistsQuery        = composable.ExistsQuery
)

type (
	CreateWarehouseSupplyRelationResult      = composable.CreateResult
	UpdateWarehouseSupplyRelationResult      = composable.MutateResult
	DeleteWarehouseSupplyRelationResult      = composable.MutateResult
	SetWarehouseSupplyRelationArchivedResult = composable.MutateResult
	GetWarehouseSupplyRelationByIdResult     = composable.GetOneResult
	SearchWarehouseSupplyRelationsResult     = composable.SearchResult
	WarehouseSupplyRelationExistsResult      = composable.ExistsResult
)
