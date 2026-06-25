package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	DriveFileShareSchemaName = "drive2.drive_file_share"

	DriveFileShareFieldId         = basemodel.FieldId
	DriveFileShareFieldFileRef    = "file_ref"
	DriveFileShareFieldUserRef    = "user_ref"
	DriveFileShareFieldPermission = "permission"

	DriveFileShareEdgeDriveFile = "drive_file"
)

func DriveFileShareSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(DriveFileShareSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Drive file share"}).
		TableName("dri_file_shares").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(DriveFileShareFieldFileRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "File"}).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(DriveFileShareFieldUserRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "User"}).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileShareFieldPermission).
				Label(model.LangJson{model.LanguageCodeEnUs: "Permission"}).
				DataType(dmodel.FieldDataTypeEnumString(driveFilePermEnumValues())).
				RequiredForCreate().
				Default(string(DriveFilePermDefault)),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(DriveFileShareEdgeDriveFile).
				Label(model.LangJson{model.LanguageCodeEnUs: "Drive file"}).
				ManyToOne(DriveFileSchemaName, dmodel.DynamicFields{
					DriveFileShareFieldFileRef: DriveFileFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type DriveFileShare struct {
	basemodel.DynamicModelBase
}

func NewDriveFileShare() *DriveFileShare {
	return &DriveFileShare{basemodel.NewDynamicModel()}
}

func NewDriveFileShareFrom(src dmodel.DynamicFields) *DriveFileShare {
	return &DriveFileShare{basemodel.NewDynamicModel(src)}
}

func (this DriveFileShare) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *DriveFileShare) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this DriveFileShare) GetFileRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileShareFieldFileRef)
}

func (this *DriveFileShare) SetFileRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileShareFieldFileRef, v)
}

func (this DriveFileShare) GetUserRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileShareFieldUserRef)
}

func (this *DriveFileShare) SetUserRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileShareFieldUserRef, v)
}

func (this DriveFileShare) GetPermission() *DriveFilePerm {
	s := this.GetFieldData().GetString(DriveFileShareFieldPermission)
	if s == nil {
		return nil
	}
	p := DriveFilePerm(*s)
	return &p
}

func (this *DriveFileShare) SetPermission(v *DriveFilePerm) {
	if v == nil {
		this.GetFieldData().SetString(DriveFileShareFieldPermission, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(DriveFileShareFieldPermission, &s)
}

func (this DriveFileShare) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(basemodel.FieldEtag)
}

func (this *DriveFileShare) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(basemodel.FieldEtag, v)
}

// DriveFileShareFile là projection tối thiểu cho API (không map DB).
type DriveFileShareFile struct {
	Id       model.Id `json:"id"`
	Name     string   `json:"name"`
	IsFolder bool     `json:"is_folder"`
}

// DriveFileShareUser là projection user cho API (không map DB).
type DriveFileShareUser struct {
	Id          model.Id `json:"id"`
	DisplayName *string  `json:"display_name,omitempty"`
	Email       *string  `json:"email,omitempty"`
	AvatarUrl   *string  `json:"avatar_url,omitempty"`
}
