package composable

import (
	"net/http"
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
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
	CrudActionBulkCreate   = CrudAction("bulk_create")
	CrudActionImport       = CrudAction("import")
)

func (this CrudAction) String() string {
	return string(this)
}

// AllCrudActions lists every built-in action.
func AllCrudActions() []CrudAction {
	return []CrudAction{
		CrudActionCreate, CrudActionUpdate, CrudActionDelete, CrudActionSetArchived,
		CrudActionGetById, CrudActionGetByUnique, CrudActionSearch, CrudActionExists,
		CrudActionGetSchema, CrudActionComputeField, CrudActionBulkCreate, CrudActionImport,
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

	// ActionTypeUploadPatch is ActionTypeUpload against an existing record: a multipart PATCH.
	//
	// It exists because a resource whose create accepts a file must be able to accept one on
	// update too, at the same ":id" path its JSON update uses. Serving that through
	// ActionTypeUpdatePatch instead would make the engine attempt a JSON bind of a multipart body,
	// which bindPayload reports as a 400 before the handler ever runs.
	ActionTypeUploadPatch = ActionType("UploadPatch")
)

// IsUpload reports whether the engine leaves this action's body to the handler. Both multipart
// types do; every other type is bound before the handler is called.
func (this ActionType) IsUpload() bool {
	return this == ActionTypeUpload || this == ActionTypeUploadPatch
}

func (this ActionType) String() string {
	return string(this)
}

func (this ActionType) IsValid() bool {
	switch this {
	case ActionTypeCreate, ActionTypeDelete, ActionTypeRead,
		ActionTypeUpdatePatch, ActionTypeUpdateReplace, ActionTypeGeneric,
		ActionTypeUpload, ActionTypeUploadPatch:
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
	case ActionTypeUpdatePatch, ActionTypeUploadPatch:
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
	BulkCreateResult   = dyn.OpResult[BulkCreateResultData]
)

// BulkCreateResultData answers bulk_create and import alike: what was written, and which rows
// were skipped and why. Skipped rows are data, not client errors, because the caller cannot
// fix them by changing the request - only by changing the file.
type BulkCreateResultData struct {
	AffectedCount int                 `json:"affected_count"`
	CreatedCount  int                 `json:"created_count"`
	UpdatedCount  int                 `json:"updated_count"`
	AffectedAt    model.ModelDateTime `json:"affected_at"`
	TotalRows     int                 `json:"total_rows"`
	ErrorCount    int                 `json:"error_count"`
	Errors        []RowError          `json:"errors"`
}

// RowError locates one rejected row. Row is 1-based over the data rows, the header excluded,
// so it matches what the user sees in a spreadsheet minus the header line.
type RowError struct {
	Row    int            `json:"row"`
	Field  string         `json:"field,omitempty"`
	Code   string         `json:"code"`
	Params map[string]any `json:"params,omitempty"`
}

// Deduplication fields. A schema declaring both gets insert-or-update semantics in bulk create.
const (
	FieldSourceSystem = "source_system"
	FieldExternalId   = "external_id"

	SourceSystemManual = "manual"
	SourceSystemImport = "import"
)

// ComputeFieldResultData is the answer of a meta/compute call: the evaluated value with the
// declared type of the field, so a client can render it without knowing the function.
type ComputeFieldResultData struct {
	DataType string `json:"data_type"`
	IsArray  bool   `json:"is_array"`
	Value    any    `json:"value"`
}
