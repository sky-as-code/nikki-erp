package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
)

const (
	FileSchemaName = c.ResourceItStorageFile

	FileFieldId         = basemodel.FieldId
	FileFieldOrgId      = basemodel.FieldOrgId
	FileFieldCreatedAt  = basemodel.FieldCreatedAt
	FileFieldEtag       = basemodel.FieldEtag
	FileFieldName       = "name"
	FileFieldMimeType   = "mime_type"
	FileFieldSizeBytes  = "size_bytes"
	FileFieldSpaceId    = "space_id"
	FileFieldObjectKey  = "object_key"
	FileFieldStatus     = "status"
	FileFieldInUseAt    = "in_use_at"
	FileFieldUploadedBy = "uploaded_by"

	// FileFieldFilePath is not a column: it is derived from the space path and answered by the
	// upload and meta APIs.
	FileFieldFilePath = "file_path"
)

//go:embed itstorage_file.json
var fileSchemaJson string

func FileSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(fileSchemaJson)
}

// File is one stored object and its metadata.
type File struct {
	basemodel.DynamicModelBase
}

func NewFileFrom(src dmodel.DynamicFields) *File {
	return &File{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this File) GetName() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(FileFieldName))
}

func (this File) GetMimeType() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(FileFieldMimeType))
}

func (this File) GetSizeBytes() int64 {
	size, _ := util.AnyToInt64(this.GetFieldData().GetAny(FileFieldSizeBytes))
	return size
}

func (this File) GetSpaceId() model.Id {
	return util.ValueOrZeroOf(this.GetFieldData().GetModelId(FileFieldSpaceId))
}

func (this File) GetObjectKey() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(FileFieldObjectKey))
}

func (this File) GetStatus() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(FileFieldStatus))
}

// IsReleased tells a file a record dropped from one never referenced: both are pending, but only
// the released one has been in use.
func (this File) IsReleased() bool {
	return this.GetStatus() == storage.FileStatusPending && this.GetFieldData()[FileFieldInUseAt] != nil
}

func (this File) GetUploadedBy() *model.Id {
	return this.GetFieldData().GetModelId(FileFieldUploadedBy)
}

// LeafName is the last segment of the file's path: its id plus the extension of its object key.
func (this File) LeafName() string {
	return util.ValueOrZeroOf(this.GetId()) + ExtensionOf(this.GetObjectKey())
}
