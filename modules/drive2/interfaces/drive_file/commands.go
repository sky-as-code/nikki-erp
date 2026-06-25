package drive_file

import (
	"io"
	"mime/multipart"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*GetDriveFileByIdQuery)(nil)
	util.Unused(req)
}

type CreateDriveFileCommand struct {
	models.DriveFile
	File       multipart.File
	FileHeader *multipart.FileHeader
}

type CreateDriveFileResult = dyn.OpResult[models.DriveFile]

type UpdateDriveFileMetadataCommand = models.DriveFile

type UpdateDriveFileResult = dyn.OpResult[dyn.MutateResultData]

type UpdateBulkDriveFileMetadataCommand []UpdateDriveFileMetadataCommand

type UpdateBulkDriveFileMetadataResult = dyn.OpResult[dyn.MutateResultData]

type UpdateDriveFileContentCommand struct {
	models.DriveFile
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
	Id     model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId model.Id `json:"-"`
	Fields []string `json:"fields" query:"fields"`
}

func (GetDriveFileByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getDriveFileByIdRequestType
}

type GetDriveFileByIdResult = dyn.OpResult[models.DriveFile]

type DownloadDriveFileResultData struct {
	Filename      string
	MineType      string
	File          io.ReadCloser
	ContentLength *int64
	ContentRange  *string
}

type DownloadDriveFileQuery struct {
	Id    model.Id
	Range string
}

type DownloadDriveFileResult = dyn.OpResult[DownloadDriveFileResultData]

type GetDriveFileByParentQuery struct {
	crud.SearchParam `json:",inline"`
	FileParentId     model.Id `json:"file_parent_id" param:"drive_file_id"`
	UserId           model.Id `json:"-"`
}

type GetDriveFileByParentResultData = dyn.PagedResultData[models.DriveFile]
type GetDriveFileByParentResult = dyn.OpResult[GetDriveFileByParentResultData]

type SearchDriveFileQuery struct {
	crud.SearchParam
	UserId model.Id `json:"-"`
}

type SearchDriveFileResultData = dyn.PagedResultData[models.DriveFile]
type SearchDriveFileResult = dyn.OpResult[SearchDriveFileResultData]

type SearchDriveFilesSharedQuery struct {
	crud.SearchParam
	UserId model.Id `json:"-"`
}

type SearchDriveFilesSharedResultData = dyn.PagedResultData[models.DriveFile]
type SearchDriveFilesSharedResult = dyn.OpResult[SearchDriveFilesSharedResultData]

type GetDriveFileAncestorsQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"-"`
}

type GetDriveFileAncestorsResult = dyn.OpResult[[]*models.DriveFile]

type GetDriveFileChildrenQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	Page        int      `json:"page" query:"page"`
	Size        int      `json:"size" query:"size"`
}

type GetDriveFileChildrenResultData = dyn.PagedResultData[*models.DriveFile]
type GetDriveFileChildrenResult = dyn.OpResult[GetDriveFileChildrenResultData]

type MoveDriveFileToTrashCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"driveFileId"`
	Etag        string   `json:"etag"`
	UserId      model.Id `json:"-"`
}

type MoveDriveFileToTrashResult = dyn.OpResult[dynamicmodel.MutateResultData]

type RestoreDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

type RestoreDriveFileResult = dyn.OpResult[models.DriveFile]

type MoveDriveFileCommand struct {
	DriveFileId   model.Id  `json:"drive_file_id" param:"drive_file_id"`
	ParentFileRef *model.Id `json:"parent_file_ref,omitempty"`
	UserId        model.Id  `json:"-"`
}

type MoveDriveFileResult = dyn.OpResult[models.DriveFile]

type DeleteDriveFileCommand struct {
	DriveFileId model.Id `json:"drive_file_id" param:"driveFileId"`
	UserId      model.Id `json:"-"`
}

type DeleteDriveFileResult = dyn.OpResult[dyn.MutateResultData]

type DeleteDriveFilesCommand struct {
	DriveFileIds []model.Id `json:"drive_file_ids"`
}

type DeleteDriveFilesResult = dyn.OpResult[dyn.MutateResultData]
