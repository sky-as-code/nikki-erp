package app

import (
	"github.com/sky-as-code/nikki-erp/common/deps_inject"
	drive_file "github.com/sky-as-code/nikki-erp/modules/drive/app/drive_file_service_impl"
	drive_file_share "github.com/sky-as-code/nikki-erp/modules/drive/app/drive_file_share_service_impl"
)

func InitServices() error {
	return deps_inject.Register(
		drive_file.NewDriveFileService,
		drive_file_share.NewDriveFileShareService,
	)
}
