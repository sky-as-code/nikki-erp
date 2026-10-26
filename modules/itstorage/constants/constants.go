package constants

import core "github.com/sky-as-code/nikki-erp/modules/core/constants"

const ItStorageModuleName = "itstorage"

// UploadMaxFileSizeBytes caps one upload. It must also exist in config.default.yaml of both trees.
const UploadMaxFileSizeBytes core.ConfigName = "ITSTORAGE.UPLOAD.MAX_FILE_SIZE_BYTES"

// DefaultUploadMaxFileSizeBytes applies when the key is absent (100 MiB).
const DefaultUploadMaxFileSizeBytes int64 = 100 << 20

// MebiByte converts settings_quota.itstorage_capacity_mibytes to bytes.
const MebiByte int64 = 1 << 20

// IAM resource codes served by this module. The dynamic resource engine asserts permissions
// using the schema name as the resource code, so these must equal the XSchemaName constants
// in domain/models verbatim.
const (
	ResourceItStorageSpace = "itstorage_space"
	ResourceItStorageFile  = "itstorage_file"
)

// QuotaNameItStorage is the settings_quota name whose itstorage_capacity_mibytes caps a tenant's storage.
const QuotaNameItStorage = "itstorage"

const (
	OwnerTypeTenant = "tenant"
	OwnerTypeOrg    = "org"
	OwnerTypeUser   = "user"
)

const ManagedByTypeSystem = "system"

// REST contract under /v1/itstorage, shared with the Web BFF.
const (
	RouteFileUpload   = "/files/upload"
	RouteFileMeta     = "/files/meta"
	RouteFileDownload = "/files/download"
	// FieldFilePart is the multipart part carrying the binary; it must be the last part.
	FieldFilePart = "file"
)
