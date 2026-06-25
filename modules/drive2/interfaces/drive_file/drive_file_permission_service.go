package drive_file

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

type PermissionDomainService interface {
	ResolvePermission(ctx corectx.Context, file *models.DriveFile, userId model.Id) (models.Permission, error)
	ResolvePermissionsBatch(ctx corectx.Context, files []*models.DriveFile, userId model.Id) (
		map[model.Id]models.Permission, error)
	EnrichDriveFilesWithPermissions(ctx corectx.Context, files []*models.DriveFile, userId model.Id) error
	AssertDriveFileActionAllowed(
		ctx corectx.Context,
		file *models.DriveFile,
		userId model.Id,
		allow func(models.Permission) bool,
		vErrs *ft.ClientErrors,
	) error
}
