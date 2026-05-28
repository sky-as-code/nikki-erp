package drive_file

import (
	"mime/multipart"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
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

type GetDriveFileByIdResult = dyn.OpResult[domain.DriveFile]

type GetDriveFileByParentQuery struct {
	crud.SearchParam `json:",inline"`
	FileParentId     model.Id `json:"file_parent_id" param:"drive_file_id"`
	UserId           model.Id `json:"-"`
}

type GetDriveFileByParentResultData = dyn.PagedResultData[domain.DriveFile]
type GetDriveFileByParentResult = dyn.OpResult[GetDriveFileByParentResultData]

type SearchDriveFileQuery struct {
	crud.SearchParam
	UserId model.Id `json:"-"`
}

type SearchDriveFileResultData = dyn.PagedResultData[domain.DriveFile]
type SearchDriveFileResult = dyn.OpResult[SearchDriveFileResultData]

type SearchDriveFilesSharedQuery struct {
	crud.SearchParam
	UserId model.Id `json:"-"`
}

type SearchDriveFilesSharedResultData = dyn.PagedResultData[domain.DriveFile]
type SearchDriveFilesSharedResult = dyn.OpResult[SearchDriveFilesSharedResultData]

type GetDriveFileAncestorsQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

type GetDriveFileAncestorsResultData = []domain.DriveFile
type GetDriveFileAncestorsResult = dyn.OpResult[GetDriveFileAncestorsResultData]

type MoveDriveFileToTrashCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

type MoveDriveFileToTrashResult = dyn.OpResult[domain.DriveFile]

type RestoreDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

type RestoreDriveFileResult = dyn.OpResult[domain.DriveFile]

type MoveDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

type MoveDriveFileResult = dyn.OpResult[domain.DriveFile]

type DeleteDriveFileCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

type DeleteDriveFileResult = dyn.OpResult[dyn.MutateResultData]
