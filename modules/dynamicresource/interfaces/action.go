package interfaces

import (
	"net/http"
	"regexp"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// Built-in action names. The action map is per-engine, so plain verbs cannot collide
// across resources. When a globally unique identity is needed, use "{resourceName}.{actionName}".
const (
	ActionCreate      = "create"
	ActionUpdate      = "update"
	ActionDelete      = "delete"
	ActionSetArchived = "set_archived"
	ActionGetById     = "get_by_id"
	ActionGetByUnique = "get_by_unique"
	ActionSearch      = "search"
	ActionExists      = "exists"
	ActionGetSchema   = "get_schema"
	// ActionComputeField evaluates one function-kind computed field against an unsaved model.
	ActionComputeField = "compute_field"
)

// ActionType classifies an action for the REST engine, which maps it to an HTTP method.
// It is mandatory as soon as an action declares a RestPath, and ignored otherwise.
type ActionType string

const (
	ActionTypeCreate        = ActionType("Create")
	ActionTypeDelete        = ActionType("Delete")
	ActionTypeRead          = ActionType("Read")
	ActionTypeUpdatePatch   = ActionType("UpdatePatch")
	ActionTypeUpdateReplace = ActionType("UpdateReplace")

	// ActionTypeGeneric is for actions whose semantics are none of the CRUD verbs —
	// an operation on a resource, such as "exists" or "send_invitation".
	// It maps to POST so the action may carry a request body.
	ActionTypeGeneric = ActionType("Generic")
)

func (this ActionType) String() string {
	return string(this)
}

// IsValid reports whether the value is one of the six declared action types.
func (this ActionType) IsValid() bool {
	switch this {
	case ActionTypeCreate, ActionTypeDelete, ActionTypeRead,
		ActionTypeUpdatePatch, ActionTypeUpdateReplace, ActionTypeGeneric:
		return true
	}
	return false
}

// HttpMethod maps the action type to the HTTP verb the REST engine registers it under.
// It returns an empty string for an invalid type; callers validate before registering.
func (this ActionType) HttpMethod() string {
	switch this {
	case ActionTypeCreate, ActionTypeGeneric:
		return http.MethodPost
	case ActionTypeDelete:
		return http.MethodDelete
	case ActionTypeRead:
		return http.MethodGet
	case ActionTypeUpdatePatch:
		return http.MethodPatch
	case ActionTypeUpdateReplace:
		return http.MethodPut
	}
	return ""
}

// RestPathRegex accepts slash-separated segments of [a-zA-Z0-9_], where a segment may be an
// Echo path param (":name"). Hyphens are deliberately excluded: the word separator is "_".
var RestPathRegex = regexp.MustCompile(`^:?[a-zA-Z0-9_]+(/:?[a-zA-Z0-9_]+)*$`)

// Permission action codes, matching what the IAM application services assert.
const (
	PermissionCreate      = "create"
	PermissionRead        = "read"
	PermissionUpdate      = "update"
	PermissionDelete      = "delete"
	PermissionSetArchived = "set_archived"
)

// ActionResult is the outcome of any action. The spec writes a non-generic "OpResult";
// the real type is generic, so actions use OpResult[any] and callers type-assert Data
// to the shape the action documents:
//   - dmodel.DynamicFields for single-record actions
//   - dyn.PagedResultData[dmodel.DynamicFields] for search
//   - dyn.ExistsResultData for exists
//   - dyn.MutateResultData for delete/update/set_archived
type ActionResult = dyn.OpResult[any]

// ProcessInput is handed to the main processing function of an action.
type ProcessInput struct {
	Params dmodel.DynamicFields

	// FoundModel is the record the action's KeysToFetch identified, already fetched by the
	// pipeline. It is nil when the action declares no KeysToFetch, and when the record does
	// not exist.
	//
	// It exists so that a MainProcess needing the record does not re-read a row the pipeline
	// has already read on its behalf.
	FoundModel *dmodel.DynamicFields

	// ResourceService is the engine's service subengine. A module that installed its own
	// extended service via Engine.SetResourceService type-asserts this to its own interface.
	ResourceService DynamicResourceService

	// ResourceRepository is the engine's repository subengine, for actions that need
	// direct database access without going through the service.
	ResourceRepository DynamicResourceRepository
}

// DynamicActionProcessFn is the main business processing function of an action.
type DynamicActionProcessFn func(ctx corectx.Context, input ProcessInput) (*ActionResult, error)

// The three validator hooks are aliases of the corecrud hook types, instantiated at
// *DynamicEntity — the same type argument the service passes to corecrud.Create and
// corecrud.Update. Being aliases rather than distinct named types, they assign straight
// into corecrud.CreateParam / corecrud.UpdateParam with no adapter and no conversion:
// the service reads the hook off the action definition and hands the value over as is.
//
// The hooks are executed by the crud helper the service delegates to, not by the action
// pipeline, so they see data that InjectServiceFields and schema.Validate have already
// been through.
//
// A hook that mutates the entity in place and returns the same pointer works, because the
// field map is shared. A hook wanting to *replace* the map must return a fresh
// NewDynamicEntityFrom(...): corecrud honours a returned model only when it differs from
// the one passed in.

// ActionBeforeValidationFn may sanitize or enrich the model before schema validation.
type ActionBeforeValidationFn = corecrud.BeforeValidationFn[*DynamicEntity]

// ActionAfterValidationFn runs after successful schema validation.
type ActionAfterValidationFn = corecrud.AfterValidationSuccessFn[*DynamicEntity]

// ActionValidateExtraFn performs validation that the schema cannot express.
//
// It is aliased to corecrud's *update* hook shape, the wider of the two, so that one type
// serves every action. On update and delete, foundModel is the stored record, fetched by
// the crud helper itself. On create there is no stored record and foundModel is nil, so a
// hook that reads it must nil-check first.
type ActionValidateExtraFn = corecrud.UpdateValidateExtraFn[*DynamicEntity]

// KeysToFetchFn returns the primary or unique keys identifying the record the engine
// should auto-fetch, to be handed to MainProcess as ProcessInput.FoundModel.
//
// It does not feed ValidateExtra: on update and delete the crud helper fetches the stored
// record itself. Declare it only for an action whose MainProcess reads the record.
type KeysToFetchFn func(params dmodel.DynamicFields) dmodel.DynamicFields

// DynamicActionDefinition declares one action on a resource engine.
type DynamicActionDefinition struct {
	// ActionName is mandatory and unique within the engine.
	ActionName string

	// ActionType decides the HTTP method the REST engine registers this action under.
	// Mandatory as soon as RestPath is set, and ignored when the action is not exposed.
	ActionType ActionType

	// RestPath is the route path relative to the engine's RoutePath(), and may carry Echo
	// path params (":id"). An empty string means the resource base path itself.
	// Segments are [a-zA-Z0-9_]: the word separator is "_", hyphens are rejected.
	// Leave both this and ActionType unset to keep the action off the REST surface.
	RestPath string

	// RestHandler optionally replaces the generic handler. When set, it owns request
	// binding and response shaping, and the engine only registers the route for it.
	RestHandler echo.HandlerFunc

	// ParamSchema is optional. When provided, the pipeline validates params against it
	// before MainProcess runs. It does not gate the validator hooks below, which the crud
	// helper runs against the resource's own schema.
	ParamSchema func() *dmodel.ModelSchema

	// ValidateAsEdit runs the schema validation in "for edit" mode, which skips absent
	// fields and no-update fields. Set it for partial-update actions.
	ValidateAsEdit bool

	// KeysToFetch is optional. When provided, the engine fetches the identified record
	// and hands it to ValidateExtra as foundModel.
	KeysToFetch KeysToFetchFn

	// Permission is the action code to assert, e.g. "read", "create".
	// An empty string skips the permission check.
	Permission string

	// PermissionScope overrides the engine's default scope for this action only.
	PermissionScope *requestguard.ResourceScope

	// IsOrgScoped confines the action to one organization: the REST caller must supply
	// "?org_id=", the value must name an org the caller belongs to, and the action only ever
	// sees records of that org.
	//
	// A nil value means true. Org scoping is the default because the unsafe direction is the
	// silent one: an action that forgot to declare it would otherwise expose every org's rows
	// to anyone holding a grant. Opt out with util.ToPtr(false), and only for a resource that
	// genuinely has no owning org — schema metadata, or a resource with no org column at all.
	//
	// It has no effect on a resource whose schema declares no org_id field; such a resource
	// cannot be org-filtered and is left alone.
	IsOrgScoped *bool

	// PrimarySchema names the parent resource this action hangs off, and nests its REST route
	// under it:
	//
	//	/{PrimarySchema}/:{PrimaryRestIdParam}/{engine.RoutePath()}/{RestPath}
	//
	// so that the full path of a nested get-by-id reads
	// "/{primary-schema}/{primary-id}/{current-schema}/{current-id}".
	// Leave it nil for a top-level resource, which is the common case.
	PrimarySchema *string

	// PrimaryRestIdParam names the path parameter carrying the parent id, and is mandatory
	// whenever PrimarySchema is set. The value lands in the action params under this name.
	// Segments follow RestPathRegex: [a-zA-Z0-9_], the word separator is "_".
	PrimaryRestIdParam *string

	// The validator hooks. The definition is the single place they live: the service reads
	// them from here and passes them to the crud helper unchanged. There is no per-call way
	// to supply one, so what runs on a create is answerable by reading this definition alone.
	//
	// ModifyAction replaces these fields rather than chaining them, so a module attaching a
	// guard to an action that may already have one reads the existing hook first and calls it
	// from its own — see rejectArchivedOnCreate in inventory/dynamicengines.
	BeforeValidation       ActionBeforeValidationFn
	AfterValidationSuccess ActionAfterValidationFn
	ValidateExtra          ActionValidateExtraFn

	// MainProcess is mandatory.
	MainProcess DynamicActionProcessFn
}

// OrgScoped resolves IsOrgScoped, defaulting an unset field to true.
// Every call site asks through this method rather than reading the pointer, so the
// "nil means org-scoped" default cannot be forgotten in one place and honoured in another.
func (this DynamicActionDefinition) OrgScoped() bool {
	return this.IsOrgScoped == nil || *this.IsOrgScoped
}

// IsNested reports whether this action's REST route hangs off a parent resource.
func (this DynamicActionDefinition) IsNested() bool {
	return this.PrimarySchema != nil && *this.PrimarySchema != ""
}

// DynamicActionDelta overrides fields of an already defined action.
// Every field except ActionName is optional; a nil field keeps the existing value.
type DynamicActionDelta struct {
	// ActionName is mandatory and must name an existing action.
	ActionName string

	// ActionType, RestPath and RestHandler are plain values: a zero value keeps the existing
	// one. Withdrawing an action from the REST surface is therefore not expressible through
	// a delta — routes are registered once at startup, so there is nothing to withdraw from.
	ActionType  ActionType
	RestPath    string
	RestHandler echo.HandlerFunc

	ParamSchema func() *dmodel.ModelSchema

	// ValidateAsEdit is a pointer so that overriding it back to false is expressible.
	ValidateAsEdit *bool

	KeysToFetch KeysToFetchFn

	// Permission is a pointer so that overriding it to "" (skip the check) is expressible.
	Permission *string

	PermissionScope *requestguard.ResourceScope

	// IsOrgScoped is a pointer for the same reason it is one on the definition: nil means
	// "keep what the action already declared", and util.ToPtr(false) withdraws org scoping.
	IsOrgScoped *bool

	PrimarySchema      *string
	PrimaryRestIdParam *string

	BeforeValidation       ActionBeforeValidationFn
	AfterValidationSuccess ActionAfterValidationFn
	ValidateExtra          ActionValidateExtraFn
	MainProcess            DynamicActionProcessFn
}
