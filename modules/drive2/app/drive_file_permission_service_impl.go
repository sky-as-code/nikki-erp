package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

type DriveFilePermissionServiceImpl struct{}

func NewDriveFilePermissionService() it.DriveFilePermissionService {
	return &DriveFilePermissionServiceImpl{}
}

func (this *DriveFilePermissionServiceImpl) ResolvePermission(
	ctx corectx.Context, file *domain.DriveFile, userId model.Id,
) (it.FilePermissionResult, error) {
	panic("drive_file_permission_service: ResolvePermission unimplemented")
}

func (this *DriveFilePermissionServiceImpl) ResolvePermissionsBatch(
	ctx corectx.Context, files []*domain.DriveFile, userId model.Id,
) (map[model.Id]it.FilePermissionResult, error) {
	panic("drive_file_permission_service: ResolvePermissionsBatch unimplemented")
}

func (this *DriveFilePermissionServiceImpl) EnrichDriveFilesWithPermissions(
	ctx corectx.Context, files []*domain.DriveFile, userId model.Id,
) error {
	panic("drive_file_permission_service: EnrichDriveFilesWithPermissions unimplemented")
}

func (this *DriveFilePermissionServiceImpl) AssertDriveFileActionAllowed(
	ctx corectx.Context,
	file *domain.DriveFile,
	userId model.Id,
	allow func(it.FilePermissionResult) bool,
	vErrs *ft.ValidationErrors,
) error {
	panic("drive_file_permission_service: AssertDriveFileActionAllowed unimplemented")
}
