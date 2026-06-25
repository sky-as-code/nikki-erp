package v1

import (
	"mime/multipart"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

type CreateDriveFileRequest struct {
	Name          string                     `form:"name"`
	IsFolder      bool                       `form:"is_folder"`
	ParentFileRef *string                    `form:"parent_file_ref"`
	Visibility    models.DriveFileVisibility `form:"visibility"`
	FileHeader    *multipart.FileHeader      `form-file:"file"`
}
type CreateDriveFileResponse = httpserver.RestCreateResponse

type UpdateDriveFileMetadataRequest = it.UpdateDriveFileMetadataCommand
type UpdateDriveFileMetadataResponse = dynamicmodel.MutateResultData

type UpdateDriveFileContentRequest struct {
	Id         string                     `param:"id"`
	Etag       string                     `form:"etag"`
	Name       string                     `form:"name"`
	Visibility models.DriveFileVisibility `form:"visibility"`
	FileHeader *multipart.FileHeader      `form-file:"file"`
}
type UpdateDriveFileContentResponse = dynamicmodel.MutateResultData

type DeleteDriveFileRequest = it.DeleteDriveFileCommand
type DeleteDriveFileResponse = httpserver.RestDeleteResponse

type MoveDriveFileToTrashRequest = it.MoveDriveFileToTrashCommand
type MoveDriveFileToTrashResponse = httpserver.RestUpdateResponse

type RestoreDriveFileRequest = it.RestoreDriveFileCommand
type RestoreDriveFileResponse = httpserver.RestUpdateResponse

type MoveDriveFileRequest = it.MoveDriveFileCommand
type MoveDriveFileResponse = httpserver.RestUpdateResponse

type GetDriveFileAncestorsRequest = it.GetDriveFileAncestorsQuery

type GetDriveFileByIdRequest = it.GetDriveFileByIdQuery

type GetDriveFileByParentRequest = it.GetDriveFileByParentQuery

type SearchDriveFileRequest = it.SearchDriveFileQuery

type SearchDriveFilesSharedRequest = it.SearchDriveFilesSharedQuery

type StreamKioskMediaRequest struct {
	Id model.Id `param:"driveFileId"`
	httpserver.FileStreamRequestBase
}
