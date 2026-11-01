package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// WarehouseSupplyRelationRepository reads and writes warehouse_supply_relation rows.
type WarehouseSupplyRelationRepository interface {
	dynamicresource.CrudRepository
}

// WarehouseSupplyRelationDomainService is the CRUD of the resource plus its own rules.
type WarehouseSupplyRelationDomainService interface {
	dynamicresource.CrudDomainService
}

// WarehouseSupplyRelationApplicationService is the authorized surface the REST handler serves.
type WarehouseSupplyRelationApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateWarehouseSupplyRelationCommand      = dynamicresource.CreateCommand
	UpdateWarehouseSupplyRelationCommand      = dynamicresource.UpdateCommand
	DeleteWarehouseSupplyRelationCommand      = dynamicresource.DeleteCommand
	SetWarehouseSupplyRelationArchivedCommand = dynamicresource.SetArchivedCommand
	GetWarehouseSupplyRelationByIdQuery       = dynamicresource.GetByIdQuery
	SearchWarehouseSupplyRelationsQuery       = dynamicresource.SearchQuery
	WarehouseSupplyRelationExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateWarehouseSupplyRelationResult      = dynamicresource.CreateResult
	UpdateWarehouseSupplyRelationResult      = dynamicresource.MutateResult
	DeleteWarehouseSupplyRelationResult      = dynamicresource.MutateResult
	SetWarehouseSupplyRelationArchivedResult = dynamicresource.MutateResult
	GetWarehouseSupplyRelationByIdResult     = dynamicresource.GetOneResult
	SearchWarehouseSupplyRelationsResult     = dynamicresource.SearchResult
	WarehouseSupplyRelationExistsResult      = dynamicresource.ExistsResult
)
