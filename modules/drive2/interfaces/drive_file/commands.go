package drive_file

import (
	"mime/multipart"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

func init() {
	var req cqrs.Request
	req = (*GetDriveFileByIdQuery)(nil)
	util.Unused(req)
}

type CreateDriveFileCommand struct {
	domain.DriveFile
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type CreateDriveFileResult = dyn.OpResult[domain.DriveFile]

type UpdateDriveFileMetadataCommand struct {
	domain.DriveFile
}

type UpdateDriveFileResult = dyn.OpResult[dyn.MutateResultData]

type UpdateBulkDriveFileMetadataCommand struct {
	DriveFiles []UpdateDriveFileMetadataCommand
}

type UpdateBulkDriveFileMetadataResult = dyn.OpResult[dyn.MutateResultData]

type UpdateDriveFileContentCommand struct {
	domain.DriveFile
	File       multipart.File        `form:"-"`
	FileHeader *multipart.FileHeader `form-file:"file"`
}

type UpdateDriveFileContentResult = dyn.OpResult[dyn.MutateResultData]

var getDriveFileByIdRequestType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file",
	Action:    "getById",
}

type GetDriveFileByIdQuery struct {
	IsDownload  bool     `json:"is_download" query:"download"`
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

func (GetDriveFileByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getDriveFileByIdRequestType
}

func (this GetDriveFileByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileByIdResult = dyn.OpResult[domain.DriveFile]

type GetDriveFileByParentQuery struct {
	crud.SearchQuery `json:",inline"`
	FileParentId     model.Id `json:"file_parent_id" param:"drive_file_id"`
	UserId           model.Id `json:"-"`
}

func (this GetDriveFileByParentQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{}
	rules = append(rules, this.SearchQuery.ValidationRules()...)
	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	if this.FileParentId != "" {
		rules = append(rules, model.IdValidateRule(&this.FileParentId, true))
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this *GetDriveFileByParentQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type GetDriveFileByParentResultData = dyn.PagedResultData[domain.DriveFile]
type GetDriveFileByParentResult = dyn.OpResult[GetDriveFileByParentResultData]

type SearchDriveFileQuery struct {
	crud.SearchQuery
	UserId model.Id `json:"-"`
}

func (this SearchDriveFileQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFileQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFileResultData = dyn.PagedResultData[domain.DriveFile]
type SearchDriveFileResult = dyn.OpResult[SearchDriveFileResultData]

type SearchDriveFilesSharedQuery struct {
	crud.SearchQuery
	UserId model.Id `json:"-"`
}

func (this SearchDriveFilesSharedQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdValidateRule(&this.UserId, true))
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFilesSharedQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFilesSharedResultData = dyn.PagedResultData[domain.DriveFile]
type SearchDriveFilesSharedResult = dyn.OpResult[SearchDriveFilesSharedResultData]

type GetDriveFileAncestorsQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

func (this GetDriveFileAncestorsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileAncestorsResultData = []domain.DriveFile
type GetDriveFileAncestorsResult = dyn.OpResult[GetDriveFileAncestorsResultData]

type MoveDriveFileToTrashCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

func (this MoveDriveFileToTrashCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type MoveDriveFileToTrashResult = dyn.OpResult[domain.DriveFile]

type RestoreDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

func (this RestoreDriveFileCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(this.ParentFileRef, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type RestoreDriveFileResult = dyn.OpResult[domain.DriveFile]

type MoveDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

func (this MoveDriveFileCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(this.ParentFileRef, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type MoveDriveFileResult = dyn.OpResult[domain.DriveFile]

type DeleteDriveFileCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

func (this DeleteDriveFileCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteDriveFileResult = dyn.OpResult[dyn.MutateResultData]
