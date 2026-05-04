package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

func (this *CreateDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := domain.NewDriveFileShare()
	f := this.FileRef
	u := this.UserRef
	d.SetFileRef(&f)
	d.SetUserRef(&u)
	d.SetPermission(&this.Permission)
	return d
}

func (this CreateBulkDriveFileShareCommand) ToDomainModels() []*domain.DriveFileShare {
	models := make([]*domain.DriveFileShare, 0, len(this.UserRefs))
	for _, userRef := range this.UserRefs {
		d := domain.NewDriveFileShare()
		f := this.FileRef
		d.SetFileRef(&f)
		d.SetUserRef(&userRef)
		d.SetPermission(&this.Permission)
		models = append(models, d)
	}
	return models
}

func (this UpdateDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := domain.NewDriveFileShare()
	id := this.Id
	d.SetId(&id)
	etag := this.Etag
	d.SetEtag(&etag)
	d.SetPermission(&this.Permission)
	return d
}

func (this DeleteDriveFileShareCommand) ToDomainModel() *domain.DriveFileShare {
	d := domain.NewDriveFileShare()
	id := this.DriveFileShareId
	d.SetId(&id)
	return d
}
