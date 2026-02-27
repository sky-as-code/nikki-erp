package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

func (this *CreateDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := &domain.DriveFileShare{}
	d.FileRef = this.FileRef
	d.UserRef = this.UserRef
	d.Permission = this.Permission
	return d
}

func (this CreateBulkDriveFileShareCommand) ToDomainModels() []*domain.DriveFileShare {
	models := make([]*domain.DriveFileShare, 0, len(this.UserRefs))
	for _, userRef := range this.UserRefs {
		d := &domain.DriveFileShare{}
		d.FileRef = this.FileRef
		d.UserRef = userRef
		d.Permission = this.Permission
		models = append(models, d)
	}
	return models
}

func (this UpdateDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := &domain.DriveFileShare{}
	d.Id = (*model.Id)(&this.Id)
	d.Etag = (*model.Etag)(&this.Etag)
	d.Permission = this.Permission
	return d
}

func (this DeleteDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := &domain.DriveFileShare{}
	d.Id = (*model.Id)(&this.DriveFileShareId)
	return d
}
