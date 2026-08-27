package dynamicresource

import (
	stdErr "errors"
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/engine"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// EngineNamePrefix prefixes the dependency container name under which a feature module
// registers its engine, e.g. "engine_iam_user".
const EngineNamePrefix = "engine_"

// EngineDependencyName is the name to register an engine under, and to inject it by.
func EngineDependencyName(schemaName string) string {
	return EngineNamePrefix + schemaName
}

var registrySingleton = &engineRegistry{
	engines: map[string]it.DynamicResourceEngine{},
}

// Registry returns the process-wide engine registry.
func Registry() it.DynamicResourceEngineRegistry {
	return registrySingleton
}

// coreDeps are the services every engine needs, resolved once when this module initializes.
type coreDeps struct {
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

// initRegistryDeps resolves the core services the registry needs to build engines.
// It runs during this module's Init(), which the dependency graph places before the
// Init() of every feature module that declares this module in its Deps().
func initRegistryDeps() error {
	return deps.Invoke(func(
		client orm.DbClient,
		configSvc config.ConfigService,
		queryBuilder orm.QueryBuilder,
		logger logging.LoggerService,
		newBaseRepoFn dyn.NewBaseDynamicRepositoryFn,
	) {
		registrySingleton.setCoreDeps(coreDeps{
			Client:        client,
			ConfigSvc:     configSvc,
			QueryBuilder:  queryBuilder,
			Logger:        logger,
			NewBaseRepoFn: newBaseRepoFn,
		})
	})
}

// engineRegistry owns every resource engine of the running application.
//
// Future work: a part of this data is meant to come from the database, so that a resource
// can be declared without a code change. The seam is newEngine: a loader would read the
// persisted engine and action rows and replay them as NewEngine plus DefineAction calls.
// That loading is deliberately not implemented yet.
type engineRegistry struct {
	mutex    sync.RWMutex
	engines  map[string]it.DynamicResourceEngine
	deps     coreDeps
	depsDone bool
}

func (this *engineRegistry) setCoreDeps(deps coreDeps) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.deps = deps
	this.depsDone = true
}

// NewEngine builds an engine for the given schema, wires its three subengines and its
// built-in actions, then registers it. Call it from a feature module's Init().
// Only the first NewEngineOptions is used, the rest are ignored.
func (this *engineRegistry) NewEngine(
	schemaName string, options ...it.NewEngineOptions,
) (it.DynamicResourceEngine, error) {
	engineOpts := it.NewEngineOptions{}
	if len(options) > 0 {
		engineOpts = options[0]
	}

	schema := dmodel.GetSchema(schemaName)
	if schema == nil {
		return nil, errors.Errorf("no dynamic model schema named '%s'", schemaName)
	}

	this.mutex.Lock()
	if !this.depsDone {
		this.mutex.Unlock()
		return nil, errors.New(
			"dynamic resource registry is not initialized yet, " +
				"make sure your module declares 'dynamicresource' in its Deps()",
		)
	}
	if _, exists := this.engines[schemaName]; exists {
		this.mutex.Unlock()
		return nil, errors.Errorf("resource engine for '%s' already exists", schemaName)
	}
	coreDeps := this.deps
	this.mutex.Unlock()

	newEngine, err := buildEngine(schema, coreDeps, engineOpts)
	if err != nil {
		return nil, err
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, exists := this.engines[schemaName]; exists {
		return nil, errors.Errorf("resource engine for '%s' already exists", schemaName)
	}
	this.engines[schemaName] = newEngine
	return newEngine, nil
}

func buildEngine(
	schema *dmodel.ModelSchema, deps coreDeps, options it.NewEngineOptions,
) (it.DynamicResourceEngine, error) {
	repository := engine.NewDynamicResourceRepository(engine.NewRepositoryParam{
		Client:        deps.Client,
		ConfigSvc:     deps.ConfigSvc,
		QueryBuilder:  deps.QueryBuilder,
		Logger:        deps.Logger,
		NewBaseRepoFn: deps.NewBaseRepoFn,
		Schema:        schema,
	})
	defaultFields := options.DefaultSearchFields
	if len(defaultFields) == 0 {
		defaultFields = schema.DefaultSearchFields()
	}
	service := engine.NewDynamicResourceService(engine.NewServiceParam{
		Schema:        schema,
		Repository:    repository,
		DefaultFields: defaultFields,
		CrudActions:   options.CrudActions,
		// Deferred for the same reason invokeComputedFunction defers its own lookup: the
		// service is built before the engine that will own it exists. A closure captured here
		// also survives SetResourceService, which a field set on the engine afterwards would not.
		ActionLookup: func(actionName string) (it.DynamicActionDefinition, bool) {
			ownerEngine, ok := registrySingleton.GetEngine(schema.Name())
			if !ok {
				return it.DynamicActionDefinition{}, false
			}
			return ownerEngine.Action(actionName)
		},
	})
	// Every resource gets computed-field evaluation; a schema without computed fields passes
	// through untouched. A module's extended service embeds this wrapped one, so its overrides
	// keep layering on top.
	service = engine.WithComputedFields(
		service, searchSourceRowsForComputed, invokeComputedFunction, defaultFields)

	newEngine := engine.NewDynamicResourceEngine(engine.NewEngineParam{
		Schema:     schema,
		Repository: repository,
		Service:    service,
	})
	if err := engine.DefineBuiltinActions(newEngine, options.CrudActions...); err != nil {
		return nil, errors.Wrapf(err, "failed to define built-in actions of '%s'", schema.Name())
	}
	return newEngine, nil
}

// invokeComputedFunction runs a "function"-kind computed field's registered implementation. It
// resolves the engine by schema name for the same reason searchSourceRowsForComputed does: the
// service is built before the engine that will own it exists, so the lookup has to be deferred to
// call time.
func invokeComputedFunction(
	ctx corectx.Context, schemaName string, functionName string, fieldName string,
	rows []dmodel.DynamicFields,
) ([]any, error) {
	ownerEngine, ok := registrySingleton.GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for computed-field function '%s'", schemaName)
	}
	fn, ok := ownerEngine.ComputedFieldFunction(functionName)
	if !ok {
		// AssertComputedFunctionsDefined makes this unreachable after a successful boot; it stays
		// as a clear failure for a function registered later than the first read.
		return nil, errors.Errorf(
			"computed-field function '%s' of '%s.%s' is not registered",
			functionName, schemaName, fieldName)
	}
	return fn(ctx, it.ComputeFnRequest{
		SchemaName: schemaName,
		FieldName:  fieldName,
		Models:     rows,
	})
}

// searchSourceRowsForComputed is the batched read behind related computed fields: the rows of
// schemaName whose keyColumn is IN keys, projected down to fields. It goes through the source
// resource's own repository, so tenant/archive handling stays what a direct read would get.
func searchSourceRowsForComputed(
	ctx corectx.Context, schemaName string, keyColumn string, keys []any, fields []string,
) ([]dmodel.DynamicFields, error) {
	sourceEngine, ok := registrySingleton.GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for computed-field source '%s'", schemaName)
	}
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(keyColumn, dmodel.In, keys...)

	found, err := sourceEngine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Fields: fields,
		Graph:  graph,
		Page:   0,
		Size:   len(keys),
	})
	if err != nil {
		return nil, err
	}
	if found != nil && found.ClientErrors.Count() > 0 {
		return nil, errors.Errorf(
			"computed-field source read of '%s' failed: %v", schemaName, found.ClientErrors.ToError())
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

func (this *engineRegistry) GetEngine(schemaName string) (it.DynamicResourceEngine, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	found, exists := this.engines[schemaName]
	return found, exists
}

func (this *engineRegistry) MustGetEngine(schemaName string) it.DynamicResourceEngine {
	found, exists := this.GetEngine(schemaName)
	if !exists {
		panic(errors.Errorf("no resource engine registered for '%s'", schemaName))
	}
	return found
}

func (this *engineRegistry) AllEngines() []it.DynamicResourceEngine {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	names := make([]string, 0, len(this.engines))
	for name := range this.engines {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]it.DynamicResourceEngine, 0, len(names))
	for _, name := range names {
		result = append(result, this.engines[name])
	}
	return result
}

// assertComputedFunctionsDefined checks every engine's "function"-kind computed fields against
// the functions actually registered on it. Invoked from OnAppStarted, once every module's Init()
// has run and had its chance to register.
//
// Every engine is reported rather than only the first, so one boot surfaces the whole gap.
func assertComputedFunctionsDefined() error {
	var failures []error
	for _, resourceEngine := range registrySingleton.AllEngines() {
		if err := resourceEngine.AssertComputedFunctionsDefined(); err != nil {
			failures = append(failures, err)
		}
	}
	return stdErr.Join(failures...)
}
