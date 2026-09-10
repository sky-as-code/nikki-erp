package composable

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// DynamicResourceEngineOnion is one resource's stack of layers, built once and registered in the
// dependency container as "dynengine_{schema}". A REST handler injects it by that name to reach
// the application service; a module publishes the typed layers from it for other consumers.
type DynamicResourceEngineOnion interface {
	ResourceName() string
	Schema() *dmodel.ModelSchema
	Repository() CrudRepository
	DomainService() CrudDomainService
	ApplicationService() CrudApplicationService

	ComputedFieldFunction(name string) (ComputedFieldFn, bool)
	AssertComputedFunctionsDefined() error
}

// BuildParam carries the core services every repository needs. A module's registration closure
// declares it as a constructor argument, so the container fills it:
//
//	deps.RegisterNamed(composable.EngineDependencyName(schema), func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
//		return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{SchemaName: schema, ...}, param)
//	})
type BuildParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

// DynamicResourceEngineOnionImpl declares a resource. Only SchemaName is required; every other
// field customizes one layer or one rule, and the zero value reproduces the default behaviour.
type DynamicResourceEngineOnionImpl struct {
	SchemaName string

	// The factories receive the default layer and return the module's own, which embeds it.
	// The domain default handed over is already wrapped with computed-field evaluation.
	NewRepositoryFn    func(base CrudRepository) CrudRepository
	NewDomainServiceFn func(base CrudDomainService) CrudDomainService
	NewAppServiceFn    func(base CrudApplicationService) CrudApplicationService

	// CrudActions selects which built-in actions this resource supports. Nil means all.
	CrudActions []CrudAction

	// DefaultSearchFields is the field list a search returns when it names none. Empty means the
	// schema's own default_search_fields, then every column.
	DefaultSearchFields []string

	// PermissionScope applies to every permission check. Nil means requestguard.ResourceScopeOrg.
	PermissionScope *requestguard.ResourceScope

	// IsOrgScoped is nil-means-true; false withdraws org scoping for a resource whose org_id is
	// optional and NULL means "global".
	IsOrgScoped *bool

	// RejectArchivedOnCreate refuses is_archived in a create body for a schema that has the
	// field, ahead of any rule a derived service adds.
	RejectArchivedOnCreate bool

	// ComputedFunctions implements the schema's "function"-kind computed fields, by name.
	ComputedFunctions map[string]ComputedFieldFn

	schema     *dmodel.ModelSchema
	repository CrudRepository
	domSvc     CrudDomainService
	appSvc     CrudApplicationService
}

// MustBuild is Build for a dig constructor, where a wiring mistake should stop the boot.
func MustBuild(impl *DynamicResourceEngineOnionImpl, param BuildParam) DynamicResourceEngineOnion {
	onion, err := impl.Build(param)
	if err != nil {
		panic(err)
	}
	return onion
}

// Build wires the three layers in order, handing each default to the module's factory, and
// indexes the result so computed fields of other resources can read this one.
func (this *DynamicResourceEngineOnionImpl) Build(param BuildParam) (DynamicResourceEngineOnion, error) {
	if this.SchemaName == "" {
		return nil, errors.New("DynamicResourceEngineOnionImpl.SchemaName is required")
	}
	if this.appSvc != nil {
		return nil, errors.Errorf("resource onion '%s' is already built", this.SchemaName)
	}
	schema := dmodel.GetSchema(this.SchemaName)
	if schema == nil {
		return nil, errors.Errorf("no dynamic model schema named '%s'", this.SchemaName)
	}
	this.schema = schema

	this.repository = this.buildRepository(param)
	this.domSvc = this.buildDomainService()
	this.appSvc = this.buildApplicationService()

	registerSource(this.SchemaName, this.repository, this)
	return this, nil
}

func (this *DynamicResourceEngineOnionImpl) buildRepository(param BuildParam) CrudRepository {
	repository := NewDefaultCrudRepository(NewRepositoryParam{
		Client:        param.Client,
		ConfigSvc:     param.ConfigSvc,
		QueryBuilder:  param.QueryBuilder,
		Logger:        param.Logger,
		NewBaseRepoFn: param.NewBaseRepoFn,
		Schema:        this.schema,
	})
	if this.NewRepositoryFn != nil {
		repository = this.NewRepositoryFn(repository)
	}
	return repository
}

func (this *DynamicResourceEngineOnionImpl) buildDomainService() CrudDomainService {
	defaultFields := this.DefaultSearchFields
	if len(defaultFields) == 0 {
		defaultFields = this.schema.DefaultSearchFields()
	}
	domSvc := NewDefaultDomainService(NewDomainServiceParam{
		Schema:           this.schema,
		Repository:       this.repository,
		DefaultFields:    defaultFields,
		CrudActions:      this.CrudActions,
		ComputedFunction: this.ComputedFieldFunction,
	})
	// Every resource gets computed-field evaluation; a schema without computed fields passes
	// through untouched. The module's derived service embeds this wrapped one, so its overrides
	// keep layering on top.
	domSvc = WithComputedFields(domSvc, searchSourceRows, this.invokeComputedFunction, defaultFields)
	if this.RejectArchivedOnCreate {
		domSvc = withCreateGuard(domSvc, RejectArchivedOnCreate)
	}
	if this.NewDomainServiceFn != nil {
		domSvc = this.NewDomainServiceFn(domSvc)
	}
	return domSvc
}

func (this *DynamicResourceEngineOnionImpl) buildApplicationService() CrudApplicationService {
	appSvc := NewDefaultApplicationService(NewAppServiceParam{
		DomainService:   this.domSvc,
		PermissionScope: this.PermissionScope,
		IsOrgScoped:     this.IsOrgScoped,
		CrudActions:     this.CrudActions,
	})
	if this.NewAppServiceFn != nil {
		appSvc = this.NewAppServiceFn(appSvc)
	}
	return appSvc
}

func (this *DynamicResourceEngineOnionImpl) invokeComputedFunction(
	ctx corectx.Context, schemaName string, functionName string, fieldName string,
	rows []dmodel.DynamicFields,
) ([]any, error) {
	fn, ok := this.ComputedFieldFunction(functionName)
	if !ok {
		// AssertComputedFunctionsDefined makes this unreachable after a successful boot.
		return nil, errors.Errorf(
			"computed-field function '%s' of '%s.%s' is not registered",
			functionName, schemaName, fieldName)
	}
	return fn(ctx, ComputeFnRequest{
		SchemaName: schemaName,
		FieldName:  fieldName,
		Models:     rows,
	})
}

func (this *DynamicResourceEngineOnionImpl) ResourceName() string {
	return this.SchemaName
}

func (this *DynamicResourceEngineOnionImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

func (this *DynamicResourceEngineOnionImpl) Repository() CrudRepository {
	return this.repository
}

func (this *DynamicResourceEngineOnionImpl) DomainService() CrudDomainService {
	return this.domSvc
}

func (this *DynamicResourceEngineOnionImpl) ApplicationService() CrudApplicationService {
	return this.appSvc
}

func (this *DynamicResourceEngineOnionImpl) ComputedFieldFunction(name string) (ComputedFieldFn, bool) {
	fn, ok := this.ComputedFunctions[name]
	return fn, ok && fn != nil
}

func (this *DynamicResourceEngineOnionImpl) AssertComputedFunctionsDefined() error {
	return assertComputedFunctionsDefined(this.SchemaName, this.ComputedFieldFunction)
}
