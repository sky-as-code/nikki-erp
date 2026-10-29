package warehouse

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// InventoryLocationRepository reads and writes inventory_location rows.
type InventoryLocationRepository interface {
	composable.CrudRepository
}

// InventoryLocationDomainService is the CRUD of the resource plus its own rules.
type InventoryLocationDomainService interface {
	composable.CrudDomainService

	CreateBulk(
		ctx corectx.Context, cmds []CreateInventoryLocationCommand,
	) (*CreateInventoryLocationsResult, error)

	// CreateManagedBulk is CreateBulk keeping each command's managed_by_type and managed_by_id,
	// for a module provisioning the locations its records manage. No REST route reaches it.
	CreateManagedBulk(
		ctx corectx.Context, cmds []CreateInventoryLocationCommand,
	) (*CreateInventoryLocationsResult, error)

	// CreateManaged creates a location something manages, keeping the managed_by_type and
	// managed_by_id the command carries; Create strips them. No REST route reaches it.
	CreateManaged(
		ctx corectx.Context, cmd CreateInventoryLocationCommand,
	) (*CreateInventoryLocationResult, error)
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
	CreateInventoryLocationsResult     = dyn.OpResult[[]dmodel.DynamicFields]
	UpdateInventoryLocationResult      = composable.MutateResult
	DeleteInventoryLocationResult      = composable.MutateResult
	SetInventoryLocationArchivedResult = composable.MutateResult
	GetInventoryLocationByIdResult     = composable.GetOneResult
	SearchInventoryLocationsResult     = composable.SearchResult
	InventoryLocationExistsResult      = composable.ExistsResult
)
