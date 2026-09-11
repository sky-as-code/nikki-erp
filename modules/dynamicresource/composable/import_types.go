package composable

import (
	"mime/multipart"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Multipart form parts of POST /{resource}/import.
const (
	ImportFormFile    = "file"
	ImportFormMapping = "mapping"
)

// DefaultImportLanguage fills a mapping that names no language, so a multilingual cell still
// lands under a real language key.
const DefaultImportLanguage = "en-US"

// ImportMapping is the JSON carried by the "mapping" form part: which file column feeds which
// schema field, in which language a free-text cell is written, and whether a reference cell
// naming no stored record may create one.
type ImportMapping struct {
	LanguageCode            string          `json:"language_code"`
	CreateMissingReferences bool            `json:"create_missing_references"`
	Columns                 []ColumnMapping `json:"columns"`
}

// ColumnMapping pairs a file header with a target field name. The referenced schema of an
// edge field is derived from the schema relations, never sent by the client.
type ColumnMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ImportCommand is what the REST handler hands to the application service. Params carries
// org_id (and anything else the query string named) so the org resolves exactly as on create.
type ImportCommand struct {
	Params     dmodel.DynamicFields
	FileHeader *multipart.FileHeader
	Mapping    ImportMapping
}

// Request-level error keys: the caller can fix these by changing the request.
const (
	ErrImportFileRequired       = "err_import_file_required"
	ErrImportFileUnreadable     = "err_import_file_unreadable"
	ErrImportTooManyRows        = "err_import_too_many_rows"
	ErrImportMappingInvalid     = "err_import_mapping_invalid"
	ErrImportMappingUnknown     = "err_import_mapping_unknown_field"
	ErrImportMappingDuplicate   = "err_import_mapping_duplicate_target"
	ErrImportHeaderNotFound     = "err_import_header_not_found"
	ErrImportMandatoryUnmapped  = "err_import_mandatory_unmapped"
	ErrImportStorageUnavailable = "err_import_storage_unavailable"
)

// Row-level error keys specific to import (the shared ones live in bulk_create.go).
const (
	ErrRowCellInvalid           = "err_import_cell_invalid"
	ErrRowReferenceNotFound     = "err_import_reference_not_found"
	ErrRowReferenceAmbiguous    = "err_import_reference_ambiguous"
	ErrRowReferenceCreateFailed = "err_import_reference_create_failed"
)
