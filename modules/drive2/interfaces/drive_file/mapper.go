package drive_file

import (
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

func (this DeleteDriveFileCommand) ToDomainModel() *models.DriveFile {
	d := models.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}

func (this MoveDriveFileToTrashCommand) ToDomainModel() *models.DriveFile {
	d := models.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}

func (this RestoreDriveFileCommand) ToDomainModel() *models.DriveFile {
	d := models.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	d.SetParentFileRef(this.ParentFileRef)
	return d
}

func (this MoveDriveFileCommand) ToDomainModel() *models.DriveFile {
	d := models.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	d.SetParentFileRef(this.ParentFileRef)
	return d
}

func (this GetDriveFileAncestorsQuery) ToDomainModel() *models.DriveFile {
	d := models.NewDriveFile()
	id := this.DriveFileId
	d.SetId(&id)
	return d
}
