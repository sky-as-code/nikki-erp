package drive_file

import (
	"mime/multipart"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

func init() {
	var req cqrs.Request
	req = (*GetDriveFileByIdQuery)(nil)

	util.Unused(req)
}

type CreateDriveFileCommand struct {
	Name               string                   `json:"name" form:"name"`
	IsFolder           bool                     `json:"isFolder" form:"isFolder"`
	ParentDriveFileRef *model.Id                `json:"parentDriveFileRef" form:"parentDriveFileRef"`
	File               multipart.File           `json:"-" form:"-"`
	FileHeader         *multipart.FileHeader    `json:"-" form:"-"`
	Visibility         enum.DriveFileVisibility `json:"visibility,omitempty" form:"visibility"`
	OwnerRef           model.Id                 `json:"-" form:"-"`
}

type CreateDriveFileResult = crud.OpResult[*domain.DriveFile]

type UpdateDriveFileMetadataCommand struct {
	Id         model.Id                 `json:"-" param:"driveFileId"`
	Etag       model.Etag               `json:"etag"`
	Name       string                   `json:"name,omitempty"`
	Visibility enum.DriveFileVisibility `json:"visibility,omitempty"`
	Status     enum.DriveFileStatus     `json:"status,omitempty"`
	Size       uint64                   `json:"-"`
}

func (this UpdateDriveFileMetadataCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.EtagValidateRule(&this.Etag, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateDriveFileMetadataResult = crud.OpResult[*domain.DriveFile]

type UpdateBulkDriveFileMetadataCommand struct {
	DriveFiles []UpdateDriveFileMetadataCommand `json:"driveFiles"`
}

type UpdateBulkDriveFileMetadataResult = crud.OpResult[[]*domain.DriveFile]

type UpdateDriveFileContentCommand struct {
	Id         model.Id              `json:"-" param:"driveFileId" form:"driveFileId"`
	Etag       model.Etag            `json:"etag" form:"etag"`
	File       multipart.File        `json:"-" form:"-"`
	FileHeader *multipart.FileHeader `json:"-" form:"-"`
}

func (this UpdateDriveFileContentCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.EtagValidateRule(&this.Etag, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateDriveFileContentResult = crud.OpResult[*domain.DriveFile]

var getDriveFileByIdRequestType = cqrs.RequestType{
	Module:    "drive",
	Submodule: "driveFile",
	Action:    "getById",
}

type GetDriveFileByIdQuery struct {
	IsDownload  bool     `query:"download"`
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
	UserId      model.Id `json:"-"`
}

func (GetDriveFileByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getDriveFileByIdRequestType
}

func (this GetDriveFileByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileByIdResult = crud.OpResult[*domain.DriveFile]

type GetDriveFileByParentQuery struct {
	crud.SearchQuery `json:",inline"`
	FileParentId     model.Id `json:"fileParentId" param:"driveFileId"`
	UserId            model.Id `json:"-"`
}

func (this GetDriveFileByParentQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{}
	rules = append(rules, this.SearchQuery.ValidationRules()...)

	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	if this.FileParentId != "" {
		rules = append(rules,
			model.IdValidateRule(&this.FileParentId, true),
		)
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (this *GetDriveFileByParentQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type GetDriveFileByParentResultData = crud.PagedResult[*domain.DriveFile]
type GetDriveFileByParentResult = crud.OpResult[*GetDriveFileByParentResultData]

type SearchDriveFileQuery struct {
	crud.SearchQuery
	UserId model.Id `json:"-"`
}

func (this SearchDriveFileQuery) Validate() fault.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFileQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFilesSharedQuery struct {
	crud.SearchQuery
	UserId model.Id `json:"-"`
}

func (this SearchDriveFilesSharedQuery) Validate() fault.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFilesSharedQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFileResultData = crud.PagedResult[*domain.DriveFile]
type SearchDriveFileResult = crud.OpResult[*SearchDriveFileResultData]

type GetDriveFileAncestorsQuery struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
	UserId      model.Id `json:"-"              `
}

func (this GetDriveFileAncestorsQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileAncestorsResultData = []*domain.DriveFile
type GetDriveFileAncestorsResult = crud.OpResult[GetDriveFileAncestorsResultData]

type MoveDriveFileToTrashCommand struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
}

func (this MoveDriveFileToTrashCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type MoveDriveFileToTrashResult = crud.OpResult[*domain.DriveFile]

type RestoreDriveFileTo struct {
}

type RestoreDriveFileCommand struct {
	DriveFileId   model.Id  `json:"driveFileId" param:"driveFileId"`
	ParentFileRef *model.Id `json:"parentFileRef,omitempty"`
}

func (this RestoreDriveFileCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(this.ParentFileRef, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type RestoreDriveFileResult = crud.OpResult[*domain.DriveFile]

type MoveDriveFileCommand struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
	ParentFileRef *model.Id `json:"parentFileRef,omitempty"`
}

func (this MoveDriveFileCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(this.ParentFileRef, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type MoveDriveFileResult = crud.OpResult[*domain.DriveFile]

type DeleteDriveFileCommand struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
}

func (this DeleteDriveFileCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteDriveFileResult = crud.DeletionResult
