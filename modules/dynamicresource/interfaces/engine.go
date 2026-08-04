package interfaces

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// DynamicResourceEngine is the generic CRUD machinery of one resource.
// A feature module creates one instance per resource in its Init(), defines extra actions
// on it, and registers it into the dependency container as "engine_{resource name}".
type DynamicResourceEngine interface {
	// ResourceName is the dynamic-model schema name this engine serves, e.g. "iam_user".
	// It doubles as the permission resource code and as the REST route path segment.
	ResourceName() string

	Schema() *dmodel.ModelSchema

	// RoutePath is the REST path segment, defaulting to ResourceName().
	RoutePath() string
	SetRoutePath(path string)

	// DefaultPermissionScope applies to actions that declare no PermissionScope.
	DefaultPermissionScope() requestguard.ResourceScope
	SetDefaultPermissionScope(scope requestguard.ResourceScope)

	RestApi() DynamicRestApi
	ResourceService() DynamicResourceService
	ResourceRepository() DynamicResourceRepository

	SetRestApi(restApi DynamicRestApi)
	SetResourceService(service DynamicResourceService)
	SetResourceRepository(repository DynamicResourceRepository)

	// DefineAction registers a new action. It fails when the action name is already taken,
	// or when a mandatory field of the definition is missing.
	DefineAction(definition DynamicActionDefinition) error

	// ModifyAction overrides fields of an already defined action.
	// It fails when the named action does not exist.
	ModifyAction(delta DynamicActionDelta) error

	// Action returns a copy of the named action definition.
	Action(actionName string) (DynamicActionDefinition, bool)

	// ActionNames lists every defined action name.
	ActionNames() []string

	// ExecuteAction runs the full pipeline of the named action:
	// permission check, validation and hooks, key fetching, extra validation, main process.
	ExecuteAction(ctx corectx.Context, actionName string, params dmodel.DynamicFields) (*ActionResult, error)
}

// DynamicRestApi exposes a resource over HTTP.
type DynamicRestApi interface {
	// RegisterRoutes adds every endpoint of the resource to the given route group.
	RegisterRoutes(route *echo.Group, middlewares ...echo.MiddlewareFunc)
}

// DynamicResourceService holds the business processing of a resource. It performs
// validation and orchestration, and invokes the repository when it needs to touch the database.
// It performs no permission check: that belongs to the engine pipeline, which sits above it.
//
// A module extends it by embedding the default implementation into its own struct and
// installing that struct with Engine.SetResourceService.
type DynamicResourceService interface {
	Create(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dmodel.DynamicFields], error)
	Update(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error)
	Delete(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error)
	SetArchived(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error)

	// GetById fetches one record by primary key. Params carry "id" and optional "fields".
	GetById(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error)

	// GetOne fetches one record by any unique key carried in params.
	GetOne(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error)

	Search(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
	Exists(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[dyn.ExistsResultData], error)

	// Schema is the dynamic-model schema this service operates on.
	Schema() *dmodel.ModelSchema
}

// DynamicResourceRepository writes and reads the resource records through the SQL query builder.
type DynamicResourceRepository interface {
	// Embedding this interface lets the repository be passed directly to the generic
	// helpers of modules/core/dynamicmodel/crud, which expect a dyn.DynamicModelRepository.
	dyn.DynamicModelRepository

	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)

	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error)
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error)

	// FindByKeys fetches the single record identified by the given primary or unique keys.
	FindByKeys(ctx corectx.Context, keys dmodel.DynamicFields) (*dyn.OpResult[dmodel.DynamicFields], error)

	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[dmodel.DynamicFields], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
	Exists(ctx corectx.Context, keys []dmodel.DynamicFields) (*dyn.OpResult[dyn.RepoExistsResult], error)
}
