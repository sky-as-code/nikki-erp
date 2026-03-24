package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	drive_file_share "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
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

func filePermissionResult(isFolder bool, perm enum.DriveFilePerm) FilePermissionResult {
	return FilePermissionResult{IsFolder: isFolder, Permission: perm}
}

// storedSharePerm keeps only DB-backed share levels (view / edit / edit-trash).
func storedSharePerm(p enum.DriveFilePerm) enum.DriveFilePerm {
	switch p {
	case enum.DriveFilePermView, enum.DriveFilePermEdit, enum.DriveFilePermEditTrash:
		return p
	default:
		return enum.DriveFilePermNone
	}
}

// maxStoredSharePermAcrossAncestors returns the strongest stored level among shares on different
// ancestor files (each file_ref + user_ref is at most one row; the slice is one row per ancestor).
func maxStoredSharePermAcrossAncestors(shares []*domain.DriveFileShare) enum.DriveFilePerm {
	max := enum.DriveFilePermNone
	for _, share := range shares {
		if share == nil {
			continue
		}
		if p := storedSharePerm(share.Permission); p > max {
			max = p
		}
	}
	return max
}

func toInheritedPerm(base enum.DriveFilePerm) enum.DriveFilePerm {
	switch base {
	case enum.DriveFilePermView:
		return enum.DriveFilePermInheritedView
	case enum.DriveFilePermEdit:
		return enum.DriveFilePermInheritedEdit
	case enum.DriveFilePermEditTrash:
		return enum.DriveFilePermInheritedEditTrash
	default:
		return enum.DriveFilePermNone
	}
}

func (this *DriveFileServiceImpl) resolvePermission(
	ctx crud.Context,
	file *domain.DriveFile,
	userId model.Id,
) (FilePermissionResult, error) {
	if file == nil || file.Id == nil {
		return filePermissionResult(false, enum.DriveFilePermNone), nil
	}

	isFolder := file.IsFolder

	if file.OwnerRef != nil && *file.OwnerRef == userId {
		return filePermissionResult(isFolder, enum.DriveFilePermOwner), nil
	}

	directShareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: []model.Id{*file.Id},
			UserId:       userId,
		},
	)
	if err != nil {
		return filePermissionResult(isFolder, enum.DriveFilePermNone), err
	}
	directPerm := enum.DriveFilePermNone
	if directShareResult != nil && directShareResult.ClientError == nil && directShareResult.HasData &&
		len(directShareResult.Data) > 0 {
		if sh := directShareResult.Data[0]; sh != nil {
			directPerm = storedSharePerm(sh.Permission)
		}
	}

	// Direct share on this node wins over inherited / ancestor-owner. Do not merge or compare
	// with ancestor resolution here — when a direct row exists it is the effective grant (creating
	// a direct share on a path that already has inherited access should require a strictly higher
	// stored level at write time).
	if directPerm != enum.DriveFilePermNone {
		return filePermissionResult(isFolder, directPerm), nil
	}

	inheritedPerm, err := this.resolvePermissionFromAncestors(ctx, *file.Id, userId)
	if err != nil {
		return filePermissionResult(isFolder, enum.DriveFilePermNone), err
	}
	if inheritedPerm != enum.DriveFilePermNone {
		return filePermissionResult(isFolder, inheritedPerm), nil
	}

	return filePermissionResult(isFolder, enum.DriveFilePermNone), nil
}

func (this *DriveFileServiceImpl) resolvePermissionFromAncestors(
	ctx crud.Context,
	driveFileId model.Id,
	userId model.Id,
) (enum.DriveFilePerm, error) {
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, driveFileId)
	if err != nil {
		return enum.DriveFilePermNone, err
	}

	ancestorIds := make([]model.Id, 0, len(ancestors))
	for _, ancestor := range ancestors {
		if ancestor == nil || ancestor.Id == nil {
			continue
		}

		ancestorIds = append(ancestorIds, *ancestor.Id)
		if ancestor.OwnerRef != nil && *ancestor.OwnerRef == userId {
			return enum.DriveFilePermAncestorOwner, nil
		}
	}
	if len(ancestorIds) == 0 {
		return enum.DriveFilePermNone, nil
	}
	shareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: ancestorIds,
			UserId:       userId,
		},
	)

	if err != nil {
		return enum.DriveFilePermNone, err
	}

	if shareResult != nil && shareResult.ClientError != nil {
		return enum.DriveFilePermNone, nil
	}

	if shareResult == nil || !shareResult.HasData || len(shareResult.Data) == 0 {
		return enum.DriveFilePermNone, nil
	}

	maxBase := maxStoredSharePermAcrossAncestors(shareResult.Data)
	if maxBase == enum.DriveFilePermNone {
		return enum.DriveFilePermNone, nil
	}

	return toInheritedPerm(maxBase), nil
}
