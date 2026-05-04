package drive_file_share

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type DriveFileShareService interface {
	CreateDriveFileShare(ctx corectx.Context, cmd CreateDriveFileShareCommand) (*CreateDriveFileShareResult, error)
	CreateBulkDriveFileShares(ctx corectx.Context, cmd CreateBulkDriveFileShareCommand) (
		*CreateBulkDriveFileShareResult, error)
	UpdateDriveFileShare(ctx corectx.Context, cmd UpdateDriveFileShareCommand) (*UpdateDriveFileShareResult, error)
	GetDriveFileShareById(ctx corectx.Context, query GetDriveFileShareByIdQuery) (*GetDriveFileShareByIdResult, error)
	GetDriveFileShareByFileId(ctx corectx.Context, query GetDriveFileShareByFileIdQuery) (
		*GetDriveFileShareByFileIdResult, error)
	GetDriveFileAncestorOwnersByFileId(ctx corectx.Context, query GetDriveFileAncestorOwnersByFileIdQuery) (
		*GetDriveFileAncestorOwnersByFileIdResult, error)
	GetDriveFileResolvedSharesByFileId(ctx corectx.Context, query GetDriveFileResolvedSharesByFileIdQuery) (
		*GetDriveFileResolvedSharesByFileIdResult, error)
	GetDriveFileUserShareDetails(ctx corectx.Context, query GetDriveFileUserShareDetailsQuery) (
		*GetDriveFileUserShareDetailsResult, error)
	GetDriveFileShareByUser(ctx corectx.Context, query GetDriveFileShareByUserQuery) (
		*GetDriveFileShareByUserResult, error)
	ListDriveFileSharesByFileRefsAndUser(ctx corectx.Context, query ListDriveFileSharesByFileRefsAndUserQuery) (
		*ListDriveFileSharesByFileRefsAndUserResult, error)
	SearchDriveFileShare(ctx corectx.Context, query SearchDriveFileShareQuery) (*SearchDriveFileShareResult, error)
	DeleteDriveFileShare(ctx corectx.Context, cmd DeleteDriveFileShareCommand) (*DeleteDriveFileShareResult, error)
}
