package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

func (this *CreateDriveFileShareCommand) ToDomainModel() *domain.DriveFile {
	driveFile := &domain.DriveFile{}
	model.MustCopy(this, driveFile)

	return driveFile
}

func (this *UpdateDriveFileShareCommand) ToDomainModel() *domain.DriveFile {
	driveFile := &domain.DriveFile{}
	model.MustCopy(this, driveFile)

	return driveFile
}
