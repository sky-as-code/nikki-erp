package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_share"
)

type DriveFileShareServiceImpl struct{}

func NewDriveFileShareService() it.DriveFileShareService {
	return &DriveFileShareServiceImpl{}
}

func (this *DriveFileShareServiceImpl) CreateDriveFileShare(
	ctx corectx.Context, cmd it.CreateDriveFileShareCommand,
) (*it.CreateDriveFileShareResult, error) {
	panic("drive_file_share_service: CreateDriveFileShare unimplemented")
}

func (this *DriveFileShareServiceImpl) CreateBulkDriveFileShares(
	ctx corectx.Context, cmd it.CreateBulkDriveFileShareCommand,
) (*it.CreateBulkDriveFileShareResult, error) {
	panic("drive_file_share_service: CreateBulkDriveFileShares unimplemented")
}

func (this *DriveFileShareServiceImpl) UpdateDriveFileShare(
	ctx corectx.Context, cmd it.UpdateDriveFileShareCommand,
) (*it.UpdateDriveFileShareResult, error) {
	panic("drive_file_share_service: UpdateDriveFileShare unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileShareById(
	ctx corectx.Context, query it.GetDriveFileShareByIdQuery,
) (*it.GetDriveFileShareByIdResult, error) {
	panic("drive_file_share_service: GetDriveFileShareById unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileShareByFileId(
	ctx corectx.Context, query it.GetDriveFileShareByFileIdQuery,
) (*it.GetDriveFileShareByFileIdResult, error) {
	panic("drive_file_share_service: GetDriveFileShareByFileId unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileAncestorOwnersByFileId(
	ctx corectx.Context, query it.GetDriveFileAncestorOwnersByFileIdQuery,
) (*it.GetDriveFileAncestorOwnersByFileIdResult, error) {
	panic("drive_file_share_service: GetDriveFileAncestorOwnersByFileId unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileResolvedSharesByFileId(
	ctx corectx.Context, query it.GetDriveFileResolvedSharesByFileIdQuery,
) (*it.GetDriveFileResolvedSharesByFileIdResult, error) {
	panic("drive_file_share_service: GetDriveFileResolvedSharesByFileId unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileUserShareDetails(
	ctx corectx.Context, query it.GetDriveFileUserShareDetailsQuery,
) (*it.GetDriveFileUserShareDetailsResult, error) {
	panic("drive_file_share_service: GetDriveFileUserShareDetails unimplemented")
}

func (this *DriveFileShareServiceImpl) GetDriveFileShareByUser(
	ctx corectx.Context, query it.GetDriveFileShareByUserQuery,
) (*it.GetDriveFileShareByUserResult, error) {
	panic("drive_file_share_service: GetDriveFileShareByUser unimplemented")
}

func (this *DriveFileShareServiceImpl) ListDriveFileSharesByFileRefsAndUser(
	ctx corectx.Context, query it.ListDriveFileSharesByFileRefsAndUserQuery,
) (*it.ListDriveFileSharesByFileRefsAndUserResult, error) {
	panic("drive_file_share_service: ListDriveFileSharesByFileRefsAndUser unimplemented")
}

func (this *DriveFileShareServiceImpl) SearchDriveFileShare(
	ctx corectx.Context, query it.SearchDriveFileShareQuery,
) (*it.SearchDriveFileShareResult, error) {
	panic("drive_file_share_service: SearchDriveFileShare unimplemented")
}

func (this *DriveFileShareServiceImpl) DeleteDriveFileShare(
	ctx corectx.Context, cmd it.DeleteDriveFileShareCommand,
) (*it.DeleteDriveFileShareResult, error) {
	panic("drive_file_share_service: DeleteDriveFileShare unimplemented")
}
