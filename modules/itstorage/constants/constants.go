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
	// ResourceItStorageConsumption is read-only: the upload, duplicate and delete paths keep it.
	ResourceItStorageConsumption = "itstorage_consumption"
)

// ObjectTagState is the S3 object tag carrying the file status. A lifecycle rule expires objects
// tagged ObjectTagState=core/storage.FileStatusPending after 24 hours.
const ObjectTagState = "itstorage-state"

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
	// RouteFileDownloadById is how this system's own clients fetch a file: by the id a record's
	// `*_file_id` field holds. The by-path routes above serve external apps only.
	RouteFileDownloadById = "/files/:id/download"
	// RouteFileMetaById answers a file's record by id. Unlike the standard itstorage_file GET it
	// needs no org_id, so a tenant-wide file (an avatar) is readable too.
	RouteFileMetaById = "/files/:id/meta"
	// FieldFilePart is the multipart part carrying the binary; it must be the last part.
	FieldFilePart = "file"
)
