package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
)

type PermissionDomainServiceImpl struct{}

func NewPermissionDomainService() it.PermissionDomainService {
	return &PermissionDomainServiceImpl{}
}

func (this *PermissionDomainServiceImpl) ResolvePermission(
	ctx corectx.Context, file *models.DriveFile, userId model.Id,
) (models.Permission, error) {
	panic("drive_file_permission_service: ResolvePermission unimplemented")
}

func (this *PermissionDomainServiceImpl) ResolvePermissionsBatch(
	ctx corectx.Context, files []*models.DriveFile, userId model.Id,
) (map[model.Id]models.Permission, error) {
	panic("drive_file_permission_service: ResolvePermissionsBatch unimplemented")
}

func (this *PermissionDomainServiceImpl) EnrichDriveFilesWithPermissions(
	ctx corectx.Context, files []*models.DriveFile, userId model.Id,
) error {
	panic("unimplemented")
}

func (this *PermissionDomainServiceImpl) AssertDriveFileActionAllowed(
	ctx corectx.Context,
	file *models.DriveFile,
	userId model.Id,
	allow func(models.Permission) bool,
	vErrs *ft.ClientErrors,
) error {
	panic("drive_file_permission_service: AssertDriveFileActionAllowed unimplemented")
}
