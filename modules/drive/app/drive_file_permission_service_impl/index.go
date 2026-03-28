package drive_file_permission_service_impl

import (
	driFileIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
	shareIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type DriveFilePermissionServiceImpl struct {
	driveFileRepo      driFileIt.DriveFileRepository
	driveFileShareRepo shareIt.DriveFileShareRepository
}

func NewDriveFilePermissionService(
	driveFileRepo driFileIt.DriveFileRepository,
	driveFileShareRepo shareIt.DriveFileShareRepository,
) driFileIt.DriveFilePermissionService {
	return &DriveFilePermissionServiceImpl{
		driveFileRepo:      driveFileRepo,
		driveFileShareRepo: driveFileShareRepo,
	}
}
