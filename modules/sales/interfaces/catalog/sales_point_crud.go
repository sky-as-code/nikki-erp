package catalog

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SalesPointRepository reads and writes sales points: the tills, kiosks and machines a channel
// sells through.
type SalesPointRepository interface {
	composable.CrudRepository
}

// SalesPointDomainService adds the lifecycle a point has beyond its rows: suspension for the
// business state and archiving for the system one, plus the history check that decides whether a
// point may be deleted outright or only retired.
type SalesPointDomainService interface {
	composable.CrudDomainService

	Suspend(ctx corectx.Context, salesPointId string) (*dyn.OpResult[dyn.MutateResultData], error)
	Activate(ctx corectx.Context, salesPointId string) (*dyn.OpResult[dyn.MutateResultData], error)
	Archive(ctx corectx.Context, salesPointId string) (*dyn.OpResult[dyn.MutateResultData], error)
	Unarchive(ctx corectx.Context, salesPointId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// AssertCreatable refuses a point on a channel that cannot take new business.
	AssertCreatable(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// FindByExternalReference resolves the point a third-party system names by its own id, which is
	// how a machine identifies itself without knowing ours.
	FindByExternalReference(ctx corectx.Context, channelId string, externalReferenceId string) (dmodel.DynamicFields, error)

	// DeleteOrArchive removes a point that never sold anything and archives one that did: sales
	// history must keep naming the point that made it.
	DeleteOrArchive(ctx corectx.Context, salesPointId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesPointApplicationService is the authorized surface the REST handler serves.
type SalesPointApplicationService interface {
	composable.CrudApplicationService

	Suspend(ctx corectx.Context, cmd SalesPointLifecycleCommand) (*SalesPointLifecycleResult, error)
	Activate(ctx corectx.Context, cmd SalesPointLifecycleCommand) (*SalesPointLifecycleResult, error)
	Archive(ctx corectx.Context, cmd SalesPointLifecycleCommand) (*SalesPointLifecycleResult, error)
	Unarchive(ctx corectx.Context, cmd SalesPointLifecycleCommand) (*SalesPointLifecycleResult, error)
}

type (
	SalesPointLifecycleCommand = composable.UpdateCommand
	SalesPointLifecycleResult  = composable.MutateResult
)

type (
	CreateSalesPointCommand      = composable.CreateCommand
	UpdateSalesPointCommand      = composable.UpdateCommand
	DeleteSalesPointCommand      = composable.DeleteCommand
	SetSalesPointArchivedCommand = composable.SetArchivedCommand
	GetSalesPointByIdQuery       = composable.GetByIdQuery
	SearchSalesPointsQuery       = composable.SearchQuery
	SalesPointExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesPointResult      = composable.CreateResult
	UpdateSalesPointResult      = composable.MutateResult
	DeleteSalesPointResult      = composable.MutateResult
	SetSalesPointArchivedResult = composable.MutateResult
	GetSalesPointByIdResult     = composable.GetOneResult
	SearchSalesPointsResult     = composable.SearchResult
	SalesPointExistsResult      = composable.ExistsResult
)
