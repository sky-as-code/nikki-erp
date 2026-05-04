package v1

import (
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	shareIt "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_share"
)

type CreateDriveFileShareRequest = shareIt.CreateDriveFileShareCommand
type CreateDriveFileShareResponse = httpserver.RestCreateResponse

type CreateBulkDriveFileShareRequest = shareIt.CreateBulkDriveFileShareCommand

type UpdateDriveFileShareRequest = shareIt.UpdateDriveFileShareCommand
type UpdateDriveFileShareResponse = httpserver.RestUpdateResponse

type GetDriveFileShareByIdRequest = shareIt.GetDriveFileShareByIdQuery

type GetDriveFileShareByFileIdRequest = shareIt.GetDriveFileShareByFileIdQuery

type GetDriveFileAncestorOwnersByFileIdRequest = shareIt.GetDriveFileAncestorOwnersByFileIdQuery

type GetDriveFileResolvedSharesByFileIdRequest = shareIt.GetDriveFileResolvedSharesByFileIdQuery

type GetDriveFileUserShareDetailsRequest = shareIt.GetDriveFileUserShareDetailsQuery

type GetDriveFileShareByUserRequest = shareIt.GetDriveFileShareByUserQuery

type SearchDriveFileShareRequest = shareIt.SearchDriveFileShareQuery

type DeleteDriveFileShareRequest = shareIt.DeleteDriveFileShareCommand
type DeleteDriveFileShareResponse = httpserver.RestDeleteResponse
