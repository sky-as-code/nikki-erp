package app

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type DriveFileShareServiceImpl struct {
	driveFileShareRepo it.DriveFileShareRepository
}

func NewDriveFileShareService(driveFileShareRepo it.DriveFileShareRepository) it.DriveFileShareService {
	return &DriveFileShareServiceImpl{
		driveFileShareRepo: driveFileShareRepo,
	}
}

func (d *DriveFileShareServiceImpl) CreateBulkDriveFileShares(ctx crud.Context, cmd it.CreateBulkDriveFileShareCommand) (*it.CreateBulkDriveFileShareResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) CreateDriveFileShare(ctx crud.Context, cmd it.CreateDriveFileShareCommand) (*it.CreateDriveFileShareResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) DeleteDriveFileShare(ctx crud.Context, cmd it.DeleteDriveFileShareCommand) (*it.DeleteDriveFileShareResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) GetDriveFileShareByFileId(ctx crud.Context, query it.GetDriveFileShareByFileIdQuery) (*it.GetDriveFileShareByFileIdResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) GetDriveFileShareById(ctx crud.Context, query it.GetDriveFileShareByIdQuery) (*it.GetDriveFileShareByIdResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) GetDriveFileShareByUser(ctx crud.Context, query it.GetDriveFileShareByUserQuery) (*it.GetDriveFileShareByUserResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) SearchDriveFileShare(ctx crud.Context, query it.SearchDriveFileShareQuery) (*it.SearchDriveFileShareResult, error) {
	panic("unimplemented")
}

func (d *DriveFileShareServiceImpl) UpdateDriveFileShare(ctx crud.Context, cmd it.UpdateDriveFileShareCommand) (*it.UpdateDriveFileShareResult, error) {
	panic("unimplemented")
}
