package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// WarehouseRepository reads and writes warehouse rows.
type WarehouseRepository interface {
	dynamicresource.CrudRepository
}

// WarehouseDomainService is the CRUD of the resource plus its own rules.
type WarehouseDomainService interface {
	dynamicresource.CrudDomainService
}

// WarehouseApplicationService is the authorized surface the REST handler serves.
type WarehouseApplicationService interface {
	dynamicresource.CrudApplicationService
	WarehouseActionService
	WarehouseAppService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateWarehouseRowCommand   = dynamicresource.CreateCommand
	UpdateWarehouseCommand      = dynamicresource.UpdateCommand
	DeleteWarehouseCommand      = dynamicresource.DeleteCommand
	SetWarehouseArchivedCommand = dynamicresource.SetArchivedCommand
	GetWarehouseByIdQuery       = dynamicresource.GetByIdQuery
	SearchWarehousesQuery       = dynamicresource.SearchQuery
	WarehouseExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateWarehouseRowResult   = dynamicresource.CreateResult
	UpdateWarehouseResult      = dynamicresource.MutateResult
	DeleteWarehouseResult      = dynamicresource.MutateResult
	SetWarehouseArchivedResult = dynamicresource.MutateResult
	GetWarehouseByIdResult     = dynamicresource.GetOneResult
	SearchWarehousesResult     = dynamicresource.SearchResult
	WarehouseExistsResult      = dynamicresource.ExistsResult
)
