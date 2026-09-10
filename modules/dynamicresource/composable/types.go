package composable

import (
	"net/http"
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// EngineNamePrefix distinguishes composable engines from package engine's "engine_" names in
// the dependency container, so both generations coexist while modules migrate.
const EngineNamePrefix = "dynengine_"

// EngineDependencyName is the container name of the onion serving the given schema.
func EngineDependencyName(schemaName string) string {
	return EngineNamePrefix + schemaName
}

// CrudAction names one built-in action. The values double as action labels in error messages
// and as the allow-list vocabulary of NewDomainServiceParam.CrudActions.
type CrudAction string

const (
	CrudActionCreate       = CrudAction("create")
	CrudActionUpdate       = CrudAction("update")
	CrudActionDelete       = CrudAction("delete")
	CrudActionSetArchived  = CrudAction("set_archived")
	CrudActionGetById      = CrudAction("get_by_id")
	CrudActionGetByUnique  = CrudAction("get_by_unique")
	CrudActionSearch       = CrudAction("search")
	CrudActionExists       = CrudAction("exists")
	CrudActionGetSchema    = CrudAction("get_schema")
	CrudActionComputeField = CrudAction("compute_field")
)

func (this CrudAction) String() string {
	return string(this)
}

// AllCrudActions lists every built-in action.
func AllCrudActions() []CrudAction {
	return []CrudAction{
		CrudActionCreate, CrudActionUpdate, CrudActionDelete, CrudActionSetArchived,
		CrudActionGetById, CrudActionGetByUnique, CrudActionSearch, CrudActionExists,
		CrudActionGetSchema, CrudActionComputeField,
	}
}

// ActionType classifies a route for the REST engine, which maps it to an HTTP method and
// decides how the request payload is bound.
type ActionType string

const (
	ActionTypeCreate        = ActionType("Create")
	ActionTypeDelete        = ActionType("Delete")
	ActionTypeRead          = ActionType("Read")
	ActionTypeUpdatePatch   = ActionType("UpdatePatch")
	ActionTypeUpdateReplace = ActionType("UpdateReplace")

	// ActionTypeGeneric is for actions whose semantics are none of the CRUD verbs, such as
	// "exists" or "suspend". It maps to POST so the action may carry a request body.
	ActionTypeGeneric = ActionType("Generic")

	// ActionTypeUpload is a multipart POST. The REST engine does not bind its body: the handler
	// reads the form itself and receives a nil payload.
	ActionTypeUpload = ActionType("Upload")
)

func (this ActionType) String() string {
	return string(this)
}

func (this ActionType) IsValid() bool {
	switch this {
	case ActionTypeCreate, ActionTypeDelete, ActionTypeRead,
		ActionTypeUpdatePatch, ActionTypeUpdateReplace, ActionTypeGeneric, ActionTypeUpload:
		return true
	}
	return false
}

// HttpMethod returns an empty string for an invalid type; callers validate before registering.
func (this ActionType) HttpMethod() string {
	switch this {
	case ActionTypeCreate, ActionTypeGeneric, ActionTypeUpload:
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

// HasRequestBody reports whether the REST engine reads a JSON body for this type. Upload is a
// multipart form, which the handler consumes itself.
func (this ActionType) HasRequestBody() bool {
	switch this {
	case ActionTypeCreate, ActionTypeGeneric, ActionTypeUpdatePatch, ActionTypeUpdateReplace:
		return true
	}
	return false
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

// Commands and queries are the field map itself. A resource declares its own aliases over these
// (type CreateUomCommand = composable.CreateCommand) so that its service signatures read as a
// hand-written module while the wire type stays the schema-agnostic map the binders, the org
// scoping helpers and the crud helpers already speak.
type (
	CreateCommand       = dmodel.DynamicFields
	UpdateCommand       = dmodel.DynamicFields
	DeleteCommand       = dmodel.DynamicFields
	SetArchivedCommand  = dmodel.DynamicFields
	GetByIdQuery        = dmodel.DynamicFields
	GetOneQuery         = dmodel.DynamicFields
	SearchQuery         = dmodel.DynamicFields
	ExistsQuery         = dmodel.DynamicFields
	GetSchemaQuery      = dmodel.DynamicFields
	ComputeFieldCommand = dmodel.DynamicFields
)

type (
	CreateResult       = dyn.OpResult[dmodel.DynamicFields]
	MutateResult       = dyn.OpResult[dyn.MutateResultData]
	GetOneResult       = dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]
	SearchResult       = dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]
	ExistsResult       = dyn.OpResult[dyn.ExistsResultData]
	GetSchemaResult    = dyn.OpResult[any]
	ComputeFieldResult = dyn.OpResult[ComputeFieldResultData]
)

// ComputeFieldResultData is the answer of a meta/compute call: the evaluated value with the
// declared type of the field, so a client can render it without knowing the function.
type ComputeFieldResultData struct {
	DataType string `json:"data_type"`
	IsArray  bool   `json:"is_array"`
	Value    any    `json:"value"`
}
