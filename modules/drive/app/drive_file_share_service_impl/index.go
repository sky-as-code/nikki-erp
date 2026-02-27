package drive_file_share_service_impl

import (
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs"
	driFileIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type DriveFileShareServiceImpl struct {
	driveFileShareRepo it.DriveFileShareRepository
	driveFileRepo      driFileIt.DriveFileRepository
	permissionSvc      driFileIt.DriveFilePermissionService
	identityCqrs       identity_cqrs.IdentityCqrsAdapter
}

func NewDriveFileShareService(
	driveFileShareRepo it.DriveFileShareRepository,
	driveFileRepo driFileIt.DriveFileRepository,
	permissionSvc driFileIt.DriveFilePermissionService,
	identityCqrsAdapter identity_cqrs.IdentityCqrsAdapter,
) it.DriveFileShareService {
	return &DriveFileShareServiceImpl{
		driveFileShareRepo: driveFileShareRepo,
		driveFileRepo:      driveFileRepo,
		permissionSvc:      permissionSvc,
		identityCqrs:       identityCqrsAdapter,
	}
}
