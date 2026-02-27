package drive_file

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

func (this *CreateDriveFileCommand) ToDomainModel() *domain.DriveFile {
	driveFile := &domain.DriveFile{}
	model.MustCopy(this, driveFile)

	return driveFile
}

func (this *UpdateDriveFileCommand) ToDomainModel() *domain.DriveFile {
	driveFile := &domain.DriveFile{}
	model.MustCopy(this, driveFile)

	return driveFile
}
