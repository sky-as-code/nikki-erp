package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type DriveFileShareService interface {
	CreateDriveFileShare(ctx crud.Context, cmd CreateDriveFileShareCommand) (
		*CreateDriveFileShareResult, error)

	CreateBulkDriveFileShares(ctx crud.Context, cmd CreateBulkDriveFileShareCommand) (
		*CreateBulkDriveFileShareResult, error)

	UpdateDriveFileShare(ctx crud.Context, cmd UpdateDriveFileShareCommand) (
		*UpdateDriveFileShareResult, error)

	GetDriveFileShareById(ctx crud.Context, query GetDriveFileShareByIdQuery) (
		*GetDriveFileShareByIdResult, error)

	GetDriveFileShareByFileId(ctx crud.Context, query GetDriveFileShareByFileIdQuery) (
		*GetDriveFileShareByFileIdResult, error)

	GetDriveFileAncestorOwnersByFileId(ctx crud.Context, query GetDriveFileAncestorOwnersByFileIdQuery) (
		*GetDriveFileAncestorOwnersByFileIdResult, error)

	GetDriveFileResolvedSharesByFileId(ctx crud.Context, query GetDriveFileResolvedSharesByFileIdQuery) (
		*GetDriveFileResolvedSharesByFileIdResult, error)

	GetDriveFileUserShareDetails(ctx crud.Context, query GetDriveFileUserShareDetailsQuery) (
		*GetDriveFileUserShareDetailsResult, error)

	GetDriveFileShareByUser(ctx crud.Context, query GetDriveFileShareByUserQuery) (
		*GetDriveFileShareByUserResult, error)

	ListDriveFileSharesByFileRefsAndUser(ctx crud.Context, query ListDriveFileSharesByFileRefsAndUserQuery) (
		*ListDriveFileSharesByFileRefsAndUserResult, error)

	SearchDriveFileShare(ctx crud.Context, query SearchDriveFileShareQuery) (
		*SearchDriveFileShareResult, error)

	DeleteDriveFileShare(ctx crud.Context, cmd DeleteDriveFileShareCommand) (
		*DeleteDriveFileShareResult, error)
}
