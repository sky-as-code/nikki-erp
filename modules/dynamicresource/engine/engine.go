package engine

import (
	"sort"
	"strings"
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

		computedFunctions: map[string]it.ComputedFieldFn{},
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
	// computedFunctions holds the Go implementations of this schema's "function"-kind computed
	// fields, keyed by the name the schema declares. See computed_functions.go.
	computedFunctions map[string]it.ComputedFieldFn
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
	if err := validateRestFields(definition); err != nil {
		return err
	}
	if err := validateNestingFields(definition); err != nil {
		return err
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, exists := this.actions[definition.ActionName]; exists {
		return errors.Errorf(
			"action '%s' is already defined on resource '%s'",
			definition.ActionName, this.ResourceName(),
		)
	}
	if err := this.assertRouteFree(definition); err != nil {
		return err
	}
	this.actions[definition.ActionName] = definition
	return nil
}

// validateRestFields checks the REST surface of a definition. An action opts into the REST
// surface by setting ActionType or RestPath; doing so makes a valid ActionType mandatory,
// because it is what decides the HTTP method.
func validateRestFields(definition it.DynamicActionDefinition) error {
	if definition.ActionType == "" && definition.RestPath == "" {
		return nil
	}
	if definition.ActionType == "" {
		return errors.Errorf(
			"action '%s' declares a RestPath and therefore requires an ActionType",
			definition.ActionName,
		)
	}
	if !definition.ActionType.IsValid() {
		return errors.Errorf(
			"action '%s' has invalid ActionType '%s'",
			definition.ActionName, definition.ActionType,
		)
	}
	if definition.RestPath == "" {
		return nil
	}
	if strings.Contains(definition.RestPath, "-") {
		return errors.Errorf(
			"action '%s' has RestPath '%s': the word separator is '_', hyphens are not allowed",
			definition.ActionName, definition.RestPath,
		)
	}
	if !it.RestPathRegex.MatchString(definition.RestPath) {
		return errors.Errorf(
			"action '%s' has malformed RestPath '%s'",
			definition.ActionName, definition.RestPath,
		)
	}
	return nil
}

// validateNestingFields checks the primary-resource fields. The two are a pair: a parent schema
// with no id param would nest under a path segment nothing fills in, and an id param with no
// parent schema names a parent that never appears in the route.
func validateNestingFields(definition it.DynamicActionDefinition) error {
	hasSchema := definition.PrimarySchema != nil && *definition.PrimarySchema != ""
	hasIdParam := definition.PrimaryRestIdParam != nil && *definition.PrimaryRestIdParam != ""

	if !hasSchema && !hasIdParam {
		return nil
	}
	if hasSchema && !hasIdParam {
		return errors.Errorf(
			"action '%s' declares PrimarySchema '%s' and therefore requires a PrimaryRestIdParam",
			definition.ActionName, *definition.PrimarySchema,
		)
	}
	if !hasSchema && hasIdParam {
		return errors.Errorf(
			"action '%s' declares PrimaryRestIdParam '%s' without a PrimarySchema",
			definition.ActionName, *definition.PrimaryRestIdParam,
		)
	}
	if !it.RestPathRegex.MatchString(*definition.PrimarySchema) {
		return errors.Errorf(
			"action '%s' has malformed PrimarySchema '%s': the word separator is '_'",
			definition.ActionName, *definition.PrimarySchema,
		)
	}
	if !it.RestPathRegex.MatchString(*definition.PrimaryRestIdParam) {
		return errors.Errorf(
			"action '%s' has malformed PrimaryRestIdParam '%s': the word separator is '_'",
			definition.ActionName, *definition.PrimaryRestIdParam,
		)
	}
	// The engine writes the ":" itself when it builds the route, so a param declared with one
	// would produce "/::kiosk_id".
	if strings.HasPrefix(*definition.PrimaryRestIdParam, ":") {
		return errors.Errorf(
			"action '%s' has PrimaryRestIdParam '%s': declare the bare name, without the ':'",
			definition.ActionName, *definition.PrimaryRestIdParam,
		)
	}
	return nil
}

// assertRouteFree rejects a second action claiming an already-taken (method, path) pair.
// Echo's Group.Add panics on a duplicate route, so catching the clash here turns a startup
// panic into a wiring error naming both actions. Callers must hold the lock.
func (this *DynamicResourceEngineImpl) assertRouteFree(definition it.DynamicActionDefinition) error {
	if definition.ActionType == "" {
		return nil
	}

	method := definition.ActionType.HttpMethod()
	// Compared as full paths rather than RestPath alone: once an action can nest under a parent
	// resource, two actions may share a RestPath and still register distinct routes.
	fullPath := this.fullRestPath(definition)
	for name, existing := range this.actions {
		if name == definition.ActionName || existing.ActionType == "" {
			continue
		}
		if existing.ActionType.HttpMethod() == method && this.fullRestPath(existing) == fullPath {
			return errors.Errorf(
				"action '%s' claims route '%s %s' already taken by action '%s' on resource '%s'",
				definition.ActionName, method, fullPath, name, this.ResourceName(),
			)
		}
	}
	return nil
}

// fullRestPath is the route an action registers under, parent prefix included.
// It mirrors what DynamicRestApiImpl.fullRestPath registers, so that the duplicate-route check
// compares exactly the paths echo will see. The two are kept in step by routeShapeTest.
func (this *DynamicResourceEngineImpl) fullRestPath(definition it.DynamicActionDefinition) string {
	base := "/" + this.RoutePath()
	if definition.IsNested() {
		base = "/" + *definition.PrimarySchema + "/:" + *definition.PrimaryRestIdParam + base
	}
	return joinRestPath(base, definition.RestPath)
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

	merged := mergeActionDelta(definition, delta)
	if err := validateRestFields(merged); err != nil {
		return err
	}
	if err := validateNestingFields(merged); err != nil {
		return err
	}
	if err := this.assertRouteFree(merged); err != nil {
		return err
	}

	this.actions[delta.ActionName] = merged
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
	if delta.ActionType != "" {
		definition.ActionType = delta.ActionType
	}
	if delta.RestPath != "" {
		definition.RestPath = delta.RestPath
	}
	if delta.RestHandler != nil {
		definition.RestHandler = delta.RestHandler
	}
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
	if delta.IsOrgScoped != nil {
		definition.IsOrgScoped = delta.IsOrgScoped
	}
	if delta.PrimarySchema != nil {
		definition.PrimarySchema = delta.PrimarySchema
	}
	if delta.PrimaryRestIdParam != nil {
		definition.PrimaryRestIdParam = delta.PrimaryRestIdParam
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
