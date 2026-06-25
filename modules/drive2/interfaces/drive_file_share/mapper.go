package drive_file_share

import (
	domainModel "github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

func (this *CreateDriveFileShareCommand) ToDomainModel() *domainModel.DriveFileShare {
	d := domainModel.NewDriveFileShare()
	f := this.FileRef
	u := this.UserRef
	d.SetFileRef(&f)
	d.SetUserRef(&u)
	d.SetPermission(&this.Permission)
	return d
}

func (this CreateBulkDriveFileShareCommand) ToDomainModels() []*domainModel.DriveFileShare {
	models := make([]*domainModel.DriveFileShare, 0, len(this.UserRefs))
	for _, userRef := range this.UserRefs {
		d := domainModel.NewDriveFileShare()
		f := this.FileRef
		d.SetFileRef(&f)
		d.SetUserRef(&userRef)
		d.SetPermission(&this.Permission)
		models = append(models, d)
	}
	return models
}

func (this UpdateDriveFileShareCommand) ToDomainModel() *domainModel.DriveFileShare {
	d := domainModel.NewDriveFileShare()
	id := this.Id
	d.SetId(&id)
	etag := this.Etag
	d.SetEtag(&etag)
	d.SetPermission(&this.Permission)
	return d
}

func (this DeleteDriveFileShareCommand) ToDomainModel() *domainModel.DriveFileShare {
	d := domainModel.NewDriveFileShare()
	id := this.DriveFileShareId
	d.SetId(&id)
	return d
}
