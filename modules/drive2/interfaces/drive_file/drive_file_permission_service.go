package drive_file

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

// FilePermissionResult is the effective permission for a drive file node for one user.
type FilePermissionResult struct {
	IsFolder   bool
	Permission domain.DriveFilePerm
}

func driveFilePermRank(p domain.DriveFilePerm) int {
	switch p {
	case domain.DriveFilePermNone:
		return 0
	case domain.DriveFilePermView:
		return 1
	case domain.DriveFilePermInheritedView:
		return 2
	case domain.DriveFilePermEdit:
		return 3
	case domain.DriveFilePermInheritedEdit:
		return 4
	case domain.DriveFilePermEditTrash:
		return 5
	case domain.DriveFilePermInheritedEditTrash:
		return 6
	case domain.DriveFilePermAncestorOwner:
		return 7
	case domain.DriveFilePermOwner:
		return 8
	default:
		return -1
	}
}

func (this FilePermissionResult) CanView() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(domain.DriveFilePermView)
}

func (this FilePermissionResult) CanCreateTo() bool {
	return this.IsFolder && driveFilePermRank(this.Permission) >= driveFilePermRank(domain.DriveFilePermEdit)
}

func (this FilePermissionResult) CanUpdate() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(domain.DriveFilePermEdit)
}

func (this FilePermissionResult) CanDelete() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(domain.DriveFilePermAncestorOwner)
}

func (this FilePermissionResult) CanMoveToTrash() bool {
	r := driveFilePermRank(this.Permission)
	if !this.IsFolder && r >= driveFilePermRank(domain.DriveFilePermEditTrash) {
		return true
	}
	return r >= driveFilePermRank(domain.DriveFilePermInheritedEditTrash)
}

func (this FilePermissionResult) CanRestore() bool {
	return this.CanMoveToTrash()
}

type DriveFilePermissionService interface {
	ResolvePermission(ctx corectx.Context, file *domain.DriveFile, userId model.Id) (FilePermissionResult, error)
	ResolvePermissionsBatch(ctx corectx.Context, files []*domain.DriveFile, userId model.Id) (
		map[model.Id]FilePermissionResult, error)
	EnrichDriveFilesWithPermissions(ctx corectx.Context, files []*domain.DriveFile, userId model.Id) error
	AssertDriveFileActionAllowed(
		ctx corectx.Context,
		file *domain.DriveFile,
		userId model.Id,
		allow func(FilePermissionResult) bool,
		vErrs *ft.ValidationErrors,
	) error
}
