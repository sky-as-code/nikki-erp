package interfaces

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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
)

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

	// ResourceService is the engine's service subengine. A module that installed its own
	// extended service via Engine.SetResourceService type-asserts this to its own interface.
	ResourceService DynamicResourceService

	// ResourceRepository is the engine's repository subengine, for actions that need
	// direct database access without going through the service.
	ResourceRepository DynamicResourceRepository
}

// DynamicActionProcessFn is the main business processing function of an action.
type DynamicActionProcessFn func(ctx corectx.Context, input ProcessInput) (*ActionResult, error)

// ActionBeforeValidationFn may sanitize or enrich params before schema validation.
// The returned map replaces the params for the rest of the pipeline.
// Only invoked when the action declares a ParamSchema.
type ActionBeforeValidationFn func(ctx corectx.Context, params dmodel.DynamicFields, vErrs *ft.ClientErrors) (dmodel.DynamicFields, error)

// ActionAfterValidationFn runs after successful schema validation.
// Only invoked when the action declares a ParamSchema.
type ActionAfterValidationFn func(ctx corectx.Context, params dmodel.DynamicFields) error

// ActionValidateExtraFn performs validation that the schema cannot express.
// It runs whether or not a ParamSchema is declared.
// foundModel is the record auto-fetched from KeysToFetch, and is nil when the action
// declares no KeysToFetch.
type ActionValidateExtraFn func(ctx corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, vErrs *ft.ClientErrors) error

// KeysToFetchFn returns the primary or unique keys identifying the record the engine
// should auto-fetch before ValidateExtra runs. Serves update-like actions.
type KeysToFetchFn func(params dmodel.DynamicFields) dmodel.DynamicFields

// DynamicActionDefinition declares one action on a resource engine.
type DynamicActionDefinition struct {
	// ActionName is mandatory and unique within the engine.
	ActionName string

	// ParamSchema is optional. When provided, params are validated against it and only
	// then are BeforeValidation and AfterValidationSuccess invoked.
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

	BeforeValidation       ActionBeforeValidationFn
	AfterValidationSuccess ActionAfterValidationFn
	ValidateExtra          ActionValidateExtraFn

	// MainProcess is mandatory.
	MainProcess DynamicActionProcessFn
}

// DynamicActionDelta overrides fields of an already defined action.
// Every field except ActionName is optional; a nil field keeps the existing value.
type DynamicActionDelta struct {
	// ActionName is mandatory and must name an existing action.
	ActionName string

	ParamSchema func() *dmodel.ModelSchema

	// ValidateAsEdit is a pointer so that overriding it back to false is expressible.
	ValidateAsEdit *bool

	KeysToFetch KeysToFetchFn

	// Permission is a pointer so that overriding it to "" (skip the check) is expressible.
	Permission *string

	PermissionScope        *requestguard.ResourceScope
	BeforeValidation       ActionBeforeValidationFn
	AfterValidationSuccess ActionAfterValidationFn
	ValidateExtra          ActionValidateExtraFn
	MainProcess            DynamicActionProcessFn
}
