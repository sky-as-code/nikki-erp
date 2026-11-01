package warehouse

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// InventoryLocationRepository reads and writes inventory_location rows.
type InventoryLocationRepository interface {
	dynamicresource.CrudRepository
}

// InventoryLocationDomainService is the CRUD of the resource plus its own rules.
type InventoryLocationDomainService interface {
	dynamicresource.CrudDomainService

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
	dynamicresource.CrudApplicationService
	InventoryLocationActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateInventoryLocationCommand      = dynamicresource.CreateCommand
	UpdateInventoryLocationCommand      = dynamicresource.UpdateCommand
	DeleteInventoryLocationCommand      = dynamicresource.DeleteCommand
	SetInventoryLocationArchivedCommand = dynamicresource.SetArchivedCommand
	GetInventoryLocationByIdQuery       = dynamicresource.GetByIdQuery
	SearchInventoryLocationsQuery       = dynamicresource.SearchQuery
	InventoryLocationExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateInventoryLocationResult      = dynamicresource.CreateResult
	CreateInventoryLocationsResult     = dyn.OpResult[[]dmodel.DynamicFields]
	UpdateInventoryLocationResult      = dynamicresource.MutateResult
	DeleteInventoryLocationResult      = dynamicresource.MutateResult
	SetInventoryLocationArchivedResult = dynamicresource.MutateResult
	GetInventoryLocationByIdResult     = dynamicresource.GetOneResult
	SearchInventoryLocationsResult     = dynamicresource.SearchResult
	InventoryLocationExistsResult      = dynamicresource.ExistsResult
)
