package app

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
	"io"
)

type DriveFileServiceImpl struct {
	driveFileRepo it.DriveFileRepository
}

func NewDriveFileService(driveFileRepo it.DriveFileRepository) it.DriveFileService {
	return &DriveFileServiceImpl{
		driveFileRepo: driveFileRepo,
	}
}

func (d *DriveFileServiceImpl) CreateDriveFile(ctx crud.Context, cmd it.CreateDriveFileCommand) (*it.CreateDriveFileResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) DeleteDriveFile(ctx crud.Context, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) DownloadDriveFile(ctx crud.Context, query it.GetDriveFileByIdQuery) (io.ReadCloser, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) GetDriveFileById(ctx crud.Context, query it.GetDriveFileByIdQuery) (*it.GetDriveFileByIdResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) GetDriveFileByParent(ctx crud.Context, query it.GetDriveFileByParentQuery) (*it.GetDriveFileByParentResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) MoveDriveFileToTrash(ctx crud.Context, cmd it.MoveDriveFileToTrashCommand) (*it.MoveDriveFileToTrashResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) SearchDriveFile(ctx crud.Context, query it.SearchDriveFileQuery) (*it.SearchDriveFileResult, error) {
	panic("unimplemented")
}

func (d *DriveFileServiceImpl) UpdateDriveFile(ctx crud.Context, cmd it.UpdateDriveFileCommand) (*it.UpdateDriveFileResult, error) {
	panic("unimplemented")
}
