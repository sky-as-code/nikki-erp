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

	// DefineComputedFieldFunction registers the Go implementation of a "function"-kind computed
	// field. The name must match the schema's computed declaration; a schema naming a function no
	// engine defines fails AssertComputedFunctionsDefined at boot rather than at first read.
	//
	// It fails when the name is empty or already registered on this engine.
	DefineComputedFieldFunction(name string, fn ComputedFieldFn) error

	// ComputedFieldFunction returns the registered implementation, if any.
	ComputedFieldFunction(name string) (ComputedFieldFn, bool)

	// AssertComputedFunctionsDefined reports every "function"-kind computed field of this
	// engine's schema whose function was never registered. Called once at startup, after modules
	// have had their chance to register — an engine is built before its module's Init body runs,
	// so this cannot be a construction-time check.
	AssertComputedFunctionsDefined() error

	// ModifyAction overrides fields of an already defined action.
	// It fails when the named action does not exist.
	ModifyAction(delta DynamicActionDelta) error

	// Action returns a copy of the named action definition.
	Action(actionName string) (DynamicActionDefinition, bool)

	// ActionNames lists every defined action name.
	ActionNames() []string

	// ExecuteAction runs the full pipeline of the named action: org scoping, permission
	// check, ParamSchema validation, key fetching, main process.
	//
	// The validator hooks are not part of it. They run inside the crud helper the resource
	// service delegates to, so they fire on a direct service call too.
	ExecuteAction(ctx corectx.Context, actionName string, params dmodel.DynamicFields) (*ActionResult, error)
}

// ComputedFieldFn produces the value of a "function"-kind computed field.
//
// It receives the whole page at once rather than one row at a time: a search returns up to a
// page of rows, and a per-row signature would turn any lookup the function performs into an N+1.
// It must return exactly one value per row in Models, in the same order; a length mismatch is an
// error, never a partial fill.
//
// The function may resolve services from the dependency container. Resolve them once, when the
// module registers the function, and close over them — resolving per call walks the DI graph on
// every read.
type ComputedFieldFn func(ctx corectx.Context, req ComputeFnRequest) ([]any, error)

// ComputeFnRequest is what a computed-field function is given.
type ComputeFnRequest struct {
	// SchemaName and FieldName identify what is being computed, so one function can serve several
	// fields or several schemas.
	SchemaName string
	FieldName  string

	// Models are the rows to compute over: a page of rows on a read, exactly one on a
	// meta/compute call, where it is the unsaved model the client posted.
	Models []dmodel.DynamicFields

	// Args carries the caller-supplied extras of a meta/compute call. Nil on a read.
	Args map[string]any
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
