package drive_file

import (
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

func (this DeleteDriveFileCommand) ToDomainModel() *domain.DriveFile {
	d := domain.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}

func (this MoveDriveFileToTrashCommand) ToDomainModel() *domain.DriveFile {
	d := domain.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}

func (this RestoreDriveFileCommand) ToDomainModel() *domain.DriveFile {
	d := domain.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	d.SetParentFileRef(this.ParentFileRef)
	return d
}

func (this MoveDriveFileCommand) ToDomainModel() *domain.DriveFile {
	d := domain.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	d.SetParentFileRef(this.ParentFileRef)
	return d
}

func (this GetDriveFileAncestorsQuery) ToDomainModel() *domain.DriveFile {
	d := domain.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}
