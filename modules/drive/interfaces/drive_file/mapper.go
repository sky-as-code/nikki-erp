package drive_file

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

func (this *CreateDriveFileCommand) ToDomainModel() *domain.DriveFile {
	driveFile := &domain.DriveFile{}
	model.MustCopy(this, driveFile)
	if driveFile.ParentDriveFileRef != nil && *driveFile.ParentDriveFileRef == "" {
		driveFile.ParentDriveFileRef = nil
	}

	driveFile.Process()

	return driveFile
}

func (this UpdateDriveFileMetadataCommand) ToDomainModel() *domain.DriveFile {
	d := &domain.DriveFile{}
	model.MustCopy(this, d)
	return d
}

func (this UpdateBulkDriveFileMetadataCommand) ToDomainModels() []*domain.DriveFile {
	result := make([]*domain.DriveFile, len(this.DriveFiles))
	for i := range this.DriveFiles {
		result[i] = this.DriveFiles[i].ToDomainModel()
		if this.UserId != "" {
			result[i].UserId = this.UserId
		}
	}
	return result
}

func (this UpdateDriveFileContentCommand) ToDomainModel() *domain.DriveFile {
	d := &domain.DriveFile{}
	model.MustCopy(this, d)
	d.Process()

	return d
}

func (this DeleteDriveFileCommand) ToDomainModel() *domain.DriveFile {
	return &domain.DriveFile{ModelBase: model.ModelBase{Id: &this.DriveFileId}}
}

func (this MoveDriveFileToTrashCommand) ToDomainModel() *domain.DriveFile {
	return &domain.DriveFile{
		ModelBase: model.ModelBase{Id: &this.DriveFileId},
		UserId:    this.UserId,
	}
}

func (this RestoreDriveFileCommand) ToDomainModel() *domain.DriveFile {
	d := &domain.DriveFile{
		ModelBase: model.ModelBase{Id: &this.DriveFileId},
		UserId:    this.UserId,
	}

	d.ParentDriveFileRef = this.ParentFileRef

	return d
}

func (this MoveDriveFileCommand) ToDomainModel() *domain.DriveFile {
	d := &domain.DriveFile{
		ModelBase:          model.ModelBase{Id: &this.DriveFileId},
		ParentDriveFileRef: this.ParentFileRef,
		UserId:             this.UserId,
	}
	return d
}

func (this GetDriveFileAncestorsQuery) ToDomainModel() *domain.DriveFile {
	return &domain.DriveFile{ModelBase: model.ModelBase{Id: &this.DriveFileId}}
}