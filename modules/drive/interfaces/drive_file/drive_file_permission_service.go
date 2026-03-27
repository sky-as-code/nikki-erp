package drive_file

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

// FilePermissionResult is the effective permission for a drive file node for one user.
type FilePermissionResult struct {
	IsFolder   bool
	Permission enum.DriveFilePerm
}

func (p FilePermissionResult) CanView() bool {
	return p.Permission >= enum.DriveFilePermView
}

func (p FilePermissionResult) CanCreateTo() bool {
	return p.IsFolder && (p.Permission >= enum.DriveFilePermEdit)
}

func (p FilePermissionResult) CanUpdate() bool {
	return p.Permission >= enum.DriveFilePermEdit
}

func (p FilePermissionResult) CanDelete() bool {
	return p.Permission >= enum.DriveFilePermAncestorOwner
}

func (p FilePermissionResult) CanMoveToTrash() bool {
	return (!p.IsFolder && p.Permission >= enum.DriveFilePermEditTrash) ||
		p.Permission >= enum.DriveFilePermInheritedEditTrash
}

func (p FilePermissionResult) CanRestore() bool {
	return (!p.IsFolder && p.Permission >= enum.DriveFilePermEditTrash) ||
		p.Permission >= enum.DriveFilePermInheritedEditTrash
}

type DriveFilePermissionService interface {
	ResolvePermission(ctx crud.Context, file *domain.DriveFile, userId model.Id) (FilePermissionResult, error)
	ResolvePermissionsBatch(ctx crud.Context, files []*domain.DriveFile, userId model.Id) (map[model.Id]FilePermissionResult, error)
	EnrichDriveFilesWithPermissions(ctx crud.Context, files []*domain.DriveFile, userId model.Id) error
	AssertDriveFileActionAllowed(
		ctx crud.Context,
		file *domain.DriveFile,
		userId model.Id,
		allow func(FilePermissionResult) bool,
		vErrs *ft.ValidationErrors,
	) error
}
