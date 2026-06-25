package models

import (
	"math"
	"regexp"
	"strings"

	"github.com/samber/lo"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/drive"
)

const (
	DriveFileSchemaName = "drive2.drive_file"

	DriveFileFieldId            = basemodel.FieldId
	DriveFileFieldOwnerRef      = "owner_ref"
	DriveFileFieldParentFileRef = "parent_file_ref"
	DriveFileFieldName          = "name"
	DriveFileFieldMime          = "mime"
	DriveFileFieldIsFolder      = "is_folder"
	DriveFileFieldSize          = "size"
	DriveFileFieldStoragePath   = "storage_path"
	DriveFileFieldStorageKey    = "storage_key"
	DriveFileFieldStorage       = "storage"
	DriveFileFieldVisibility    = "visibility"
	DriveFileFieldStatus        = "status"
	DriveFileFieldDeletedAt     = "deleted_at"

	DriveFileEdgeParent    = "parent"
	DriveFileEdgeChildren  = "children"
	DriveFileEdgeShares    = "shares"
	DriveFileEdgeAncestors = "drive_file_ancestors"
)

var (
	DriveFileAllFields = []string{
		DriveFileFieldId,
		DriveFileFieldOwnerRef,
		DriveFileFieldParentFileRef,
		DriveFileFieldName,
		DriveFileFieldMime,
		DriveFileFieldIsFolder,
		DriveFileFieldSize,
		DriveFileFieldStoragePath,
		DriveFileFieldStorageKey,
		DriveFileFieldStorage,
		DriveFileFieldVisibility,
		DriveFileFieldStatus,
		DriveFileFieldDeletedAt,
		basemodel.FieldEtag,
		basemodel.FieldCreatedAt,
		basemodel.FieldUpdatedAt,
	}
)

func DriveFileSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(DriveFileSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Drive file"}).
		TableName("dri_files").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(DriveFileFieldOwnerRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Owner"}).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(DriveFileFieldParentFileRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Parent file"}),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldMime).
				Label(model.LangJson{model.LanguageCodeEnUs: "MIME type"}).
				DataType(dmodel.FieldDataTypeString(0, 255)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldIsFolder).
				Label(model.LangJson{model.LanguageCodeEnUs: "Is folder"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldSize).
				Label(model.LangJson{model.LanguageCodeEnUs: "Size"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt64)).
				RequiredForCreate().
				Default(int64(0)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldStoragePath).
				Label(model.LangJson{model.LanguageCodeEnUs: "Storage path"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_URL_LENGTH_MAX)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldStorageKey).
				Label(model.LangJson{model.LanguageCodeEnUs: "Storage key"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_URL_LENGTH_MAX)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldStorage).
				Label(model.LangJson{model.LanguageCodeEnUs: "Storage backend"}).
				DataType(dmodel.FieldDataTypeEnumString(driveFileStorageEnumValues())),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldVisibility).
				Label(model.LangJson{model.LanguageCodeEnUs: "Visibility"}).
				DataType(dmodel.FieldDataTypeEnumString(driveFileVisibilityEnumValues())).
				RequiredForCreate().
				Default(string(DriveFileVisibilityDefault)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldStatus).
				Label(model.LangJson{model.LanguageCodeEnUs: "Status"}).
				DataType(dmodel.FieldDataTypeEnumString(driveFileStatusEnumValues())).
				RequiredForCreate().
				Default(string(DriveFileStatusDefault)),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileFieldDeletedAt).
				Label(model.LangJson{model.LanguageCodeEnUs: "Deleted at"}).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(DriveFileEdgeParent).
				Label(model.LangJson{model.LanguageCodeEnUs: "Parent"}).
				ManyToOne(DriveFileSchemaName, dmodel.DynamicFields{
					DriveFileFieldParentFileRef: DriveFileFieldId,
				}),
		).
		EdgeFrom(
			dmodel.Edge(DriveFileEdgeChildren).
				Label(model.LangJson{model.LanguageCodeEnUs: "Children"}).
				Existing(DriveFileSchemaName, DriveFileEdgeParent),
		).
		EdgeFrom(
			dmodel.Edge(DriveFileEdgeShares).
				Label(model.LangJson{model.LanguageCodeEnUs: "Shares"}).
				Existing(DriveFileShareSchemaName, DriveFileShareEdgeDriveFile),
		).
		EdgeFrom(
			dmodel.Edge(DriveFileEdgeAncestors).
				Label(model.LangJson{model.LanguageCodeEnUs: "Ancestor closure rows"}).
				Existing(DriveFileAncestorSchemaName, DriveFileAncestorEdgeDriveFile),
		)
}

// DriveFileResolvedPermission is the effective permission for the acting user; set by enrichment, not persisted.
type DriveFileResolvedPermission struct {
	Permission DriveFilePerm `json:"-"`
}

// DriveFileOwner is populated for API responses (identity user summary); not persisted.
type DriveFileOwner struct {
	Id          model.Id `json:"id"`
	DisplayName *string  `json:"display_name,omitempty"`
	Email       *string  `json:"email,omitempty"`
	AvatarUrl   *string  `json:"avatar_url,omitempty"`
}

type DriveFile struct {
	basemodel.DynamicModelBase
	ResolvedPermission *DriveFileResolvedPermission `json:"-"`
	Owner              *DriveFileOwner              `json:"owner,omitempty"`
}

func NewDriveFile() *DriveFile {
	return &DriveFile{DynamicModelBase: basemodel.NewDynamicModel()}
}

func NewDriveFileFrom(src dmodel.DynamicFields) *DriveFile {
	return &DriveFile{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

var (
	driveFileNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	driveFileReserved  = []string{"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
)

func DriveFileNameValidate(name string) *fault.ClientErrorItem {
	if !driveFileNameRegex.MatchString(name) {
		return fault.NewValidationError(DriveFileFieldName,
			fault.ErrorKey(drive.ModuleSingleton.Name(), "validation_err"),
			"name may only contain letters, numbers, dots, underscores and hyphens")

	}

	if name[len(name)-1] == '.' || name[len(name)-1] == ' ' {
		return fault.NewValidationError(DriveFileFieldName,
			fault.ErrorKey(drive.ModuleSingleton.Name(), "validation_err"),
			"name must not end with dot or space")
	}

	upper := strings.ToUpper(name)
	for _, reserved := range driveFileReserved {
		if upper == reserved {
			return fault.NewValidationError(DriveFileFieldName,
				fault.ErrorKey(drive.ModuleSingleton.Name(), "validation_err"),
				"name is a reserved Windows filename")
		}

		if len(upper) > len(reserved) &&
			strings.HasPrefix(upper, reserved) &&
			upper[len(reserved)] == '.' {
			return fault.NewValidationError(drive.ModuleSingleton.Name(),
				fault.ErrorKey(drive.ModuleSingleton.Name(), "validation_err"),
				"name is a reserved Windows filename")
		}
	}

	return nil
}

func (this DriveFile) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *DriveFile) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this DriveFile) GetOwnerRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileFieldOwnerRef)
}

func (this *DriveFile) SetOwnerRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileFieldOwnerRef, v)
}

func (this DriveFile) GetParentFileRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileFieldParentFileRef)
}

func (this *DriveFile) SetParentFileRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileFieldParentFileRef, v)
}

func (this DriveFile) GetName() *string {
	return this.GetFieldData().GetString(DriveFileFieldName)
}

func (this *DriveFile) SetName(v *string) {
	this.GetFieldData().SetString(DriveFileFieldName, v)
}

func (this DriveFile) GetMime() *string {
	return this.GetFieldData().GetString(DriveFileFieldMime)
}

func (this *DriveFile) SetMime(v *string) {
	this.GetFieldData().SetString(DriveFileFieldMime, v)
}

func (this DriveFile) GetIsFolder() *bool {
	return this.GetFieldData().GetBool(DriveFileFieldIsFolder)
}

func (this *DriveFile) SetIsFolder(v *bool) {
	this.GetFieldData().SetBool(DriveFileFieldIsFolder, v)
}

func (this DriveFile) GetSize() *int64 {
	return this.GetFieldData().GetInt64(DriveFileFieldSize)
}

func (this *DriveFile) SetSize(v *int64) {
	this.GetFieldData().SetInt64(DriveFileFieldSize, v)
}

func (this DriveFile) GetStoragePath() *string {
	return this.GetFieldData().GetString(DriveFileFieldStoragePath)
}

func (this *DriveFile) SetStoragePath(v *string) {
	this.GetFieldData().SetString(DriveFileFieldStoragePath, v)
}

func (this DriveFile) GetStorageKey() *string {
	return this.GetFieldData().GetString(DriveFileFieldStorageKey)
}

func (this *DriveFile) SetStorageKey(v *string) {
	this.GetFieldData().SetString(DriveFileFieldStorageKey, v)
}

func (this DriveFile) GetStorage() *DriveFileStorage {
	s := this.GetFieldData().GetString(DriveFileFieldStorage)
	if s == nil {
		return nil
	}
	st := DriveFileStorage(*s)
	return &st
}

func (this *DriveFile) SetStorage(v *DriveFileStorage) {
	if v == nil {
		this.GetFieldData().SetString(DriveFileFieldStorage, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(DriveFileFieldStorage, &s)
}

func (this DriveFile) GetVisibility() *DriveFileVisibility {
	s := this.GetFieldData().GetString(DriveFileFieldVisibility)
	if s == nil {
		return nil
	}
	st := DriveFileVisibility(*s)
	return &st
}

func (this *DriveFile) SetVisibility(v *DriveFileVisibility) {
	if v == nil {
		this.GetFieldData().SetString(DriveFileFieldVisibility, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(DriveFileFieldVisibility, &s)
}

func (this DriveFile) GetStatus() *DriveFileStatus {
	s := this.GetFieldData().GetString(DriveFileFieldStatus)
	if s == nil {
		return nil
	}
	st := DriveFileStatus(*s)
	return &st
}

func (this *DriveFile) SetStatus(v *DriveFileStatus) {
	if v == nil {
		this.GetFieldData().SetString(DriveFileFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(DriveFileFieldStatus, &s)
}

func (this DriveFile) GetDeletedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(DriveFileFieldDeletedAt)
}

func (this *DriveFile) SetDeletedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(DriveFileFieldDeletedAt, v)
}

func (this DriveFile) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(basemodel.FieldEtag)
}

func (this *DriveFile) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(basemodel.FieldEtag, v)
}

func (this *DriveFile) GetChildren() []*DriveFile {
	raw := this.GetFieldData().GetAny(DriveFileEdgeChildren)
	if raw == nil {
		return nil
	}

	if children, ok := raw.([]*DriveFile); ok {
		return children
	}

	fieldRows, ok := raw.([]dmodel.DynamicFields)
	if !ok {
		return nil
	}

	out := make([]*DriveFile, len(fieldRows))
	for idx := range fieldRows {
		out[idx] = NewDriveFileFrom(fieldRows[idx])
	}

	return out
}

func (this *DriveFile) SetChildren(children []*DriveFile) {
	this.GetFieldData().SetAny(DriveFileEdgeChildren, children)
}

func (d *DriveFile) BuildTree(children []*DriveFile) {
	children = append(children, d)

	childrenMap := lo.SliceToMap(children, func(driveFile *DriveFile) (model.Id, *DriveFile) {
		return lo.FromPtr(driveFile.GetId()), driveFile
	})

	for _, driveFile := range children {
		if driveFile.GetParentFileRef() != nil {
			if parent, ok := childrenMap[*driveFile.GetParentFileRef()]; ok {
				pChildren := parent.GetChildren()
				pChildren = append(pChildren, driveFile)
				parent.SetChildren(pChildren)
			}
		}
	}
}
