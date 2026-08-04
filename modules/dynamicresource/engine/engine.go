package engine

import (
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewEngineParam carries what an engine needs at construction time.
type NewEngineParam struct {
	Schema     *dmodel.ModelSchema
	Repository it.DynamicResourceRepository
	Service    it.DynamicResourceService
}

func NewDynamicResourceEngine(param NewEngineParam) it.DynamicResourceEngine {
	engine := &DynamicResourceEngineImpl{
		schema:       param.Schema,
		routePath:    param.Schema.Name(),
		defaultScope: requestguard.ResourceScopeOrg,
		repository:   param.Repository,
		service:      param.Service,
		actions:      map[string]it.DynamicActionDefinition{},
	}
	engine.restApi = NewDynamicRestApi(engine)
	return engine
}

// DynamicResourceEngineImpl is the generic CRUD machinery of one resource.
// It owns the action definitions and the three subengines that carry them out.
type DynamicResourceEngineImpl struct {
	schema       *dmodel.ModelSchema
	routePath    string
	defaultScope requestguard.ResourceScope

	restApi    it.DynamicRestApi
	service    it.DynamicResourceService
	repository it.DynamicResourceRepository

	mutex   sync.RWMutex
	actions map[string]it.DynamicActionDefinition
}

func (this *DynamicResourceEngineImpl) ResourceName() string {
	return this.schema.Name()
}

func (this *DynamicResourceEngineImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

func (this *DynamicResourceEngineImpl) RoutePath() string {
	return this.routePath
}

func (this *DynamicResourceEngineImpl) SetRoutePath(path string) {
	this.routePath = path
}

func (this *DynamicResourceEngineImpl) DefaultPermissionScope() requestguard.ResourceScope {
	return this.defaultScope
}

func (this *DynamicResourceEngineImpl) SetDefaultPermissionScope(scope requestguard.ResourceScope) {
	this.defaultScope = scope
}

func (this *DynamicResourceEngineImpl) RestApi() it.DynamicRestApi {
	return this.restApi
}

func (this *DynamicResourceEngineImpl) ResourceService() it.DynamicResourceService {
	return this.service
}

func (this *DynamicResourceEngineImpl) ResourceRepository() it.DynamicResourceRepository {
	return this.repository
}

func (this *DynamicResourceEngineImpl) SetRestApi(restApi it.DynamicRestApi) {
	this.restApi = restApi
}

func (this *DynamicResourceEngineImpl) SetResourceService(service it.DynamicResourceService) {
	this.service = service
}

func (this *DynamicResourceEngineImpl) SetResourceRepository(repository it.DynamicResourceRepository) {
	this.repository = repository
}

// DefineAction registers a new action, rejecting duplicates so that two modules cannot
// silently overwrite each other's behavior. Use ModifyAction to change an existing one.
func (this *DynamicResourceEngineImpl) DefineAction(definition it.DynamicActionDefinition) error {
	if definition.ActionName == "" {
		return errors.New("action name is required")
	}
	if definition.MainProcess == nil {
		return errors.Errorf("action '%s' requires a MainProcess function", definition.ActionName)
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, exists := this.actions[definition.ActionName]; exists {
		return errors.Errorf(
			"action '%s' is already defined on resource '%s'",
			definition.ActionName, this.ResourceName(),
		)
	}
	this.actions[definition.ActionName] = definition
	return nil
}

// ModifyAction overrides the provided fields of an existing action, leaving the rest intact.
func (this *DynamicResourceEngineImpl) ModifyAction(delta it.DynamicActionDelta) error {
	if delta.ActionName == "" {
		return errors.New("action name is required")
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	definition, exists := this.actions[delta.ActionName]
	if !exists {
		return errors.Errorf(
			"action '%s' is not defined on resource '%s'",
			delta.ActionName, this.ResourceName(),
		)
	}

	this.actions[delta.ActionName] = mergeActionDelta(definition, delta)
	return nil
}

func (this *DynamicResourceEngineImpl) Action(actionName string) (it.DynamicActionDefinition, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	definition, exists := this.actions[actionName]
	return definition, exists
}

func (this *DynamicResourceEngineImpl) ActionNames() []string {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	names := make([]string, 0, len(this.actions))
	for name := range this.actions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// mergeActionDelta returns the definition with every non-nil delta field applied.
func mergeActionDelta(
	definition it.DynamicActionDefinition, delta it.DynamicActionDelta,
) it.DynamicActionDefinition {
	if delta.ParamSchema != nil {
		definition.ParamSchema = delta.ParamSchema
	}
	if delta.ValidateAsEdit != nil {
		definition.ValidateAsEdit = *delta.ValidateAsEdit
	}
	if delta.KeysToFetch != nil {
		definition.KeysToFetch = delta.KeysToFetch
	}
	if delta.Permission != nil {
		definition.Permission = *delta.Permission
	}
	if delta.PermissionScope != nil {
		definition.PermissionScope = delta.PermissionScope
	}
	if delta.BeforeValidation != nil {
		definition.BeforeValidation = delta.BeforeValidation
	}
	if delta.AfterValidationSuccess != nil {
		definition.AfterValidationSuccess = delta.AfterValidationSuccess
	}
	if delta.ValidateExtra != nil {
		definition.ValidateExtra = delta.ValidateExtra
	}
	if delta.MainProcess != nil {
		definition.MainProcess = delta.MainProcess
	}
	return definition
}
