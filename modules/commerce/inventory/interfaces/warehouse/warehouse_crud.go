package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// WarehouseRepository reads and writes warehouse rows.
type WarehouseRepository interface {
	composable.CrudRepository
}

// WarehouseDomainService is the CRUD of the resource plus its own rules.
type WarehouseDomainService interface {
	composable.CrudDomainService
}

// WarehouseApplicationService is the authorized surface the REST handler serves.
type WarehouseApplicationService interface {
	composable.CrudApplicationService
	WarehouseActionService
	WarehouseAppService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateWarehouseRowCommand   = composable.CreateCommand
	UpdateWarehouseCommand      = composable.UpdateCommand
	DeleteWarehouseCommand      = composable.DeleteCommand
	SetWarehouseArchivedCommand = composable.SetArchivedCommand
	GetWarehouseByIdQuery       = composable.GetByIdQuery
	SearchWarehousesQuery       = composable.SearchQuery
	WarehouseExistsQuery        = composable.ExistsQuery
)

type (
	CreateWarehouseRowResult   = composable.CreateResult
	UpdateWarehouseResult      = composable.MutateResult
	DeleteWarehouseResult      = composable.MutateResult
	SetWarehouseArchivedResult = composable.MutateResult
	GetWarehouseByIdResult     = composable.GetOneResult
	SearchWarehousesResult     = composable.SearchResult
	WarehouseExistsResult      = composable.ExistsResult
)
