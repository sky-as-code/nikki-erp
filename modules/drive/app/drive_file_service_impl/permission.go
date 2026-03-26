package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/collections"
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

// resolvePermissionsBatch resolves the effective permission for each file in a
// single batch, keeping total DB round-trips to a fixed 3-4 regardless of len(files).
func (this *DriveFileServiceImpl) resolvePermissionsBatch(
	ctx crud.Context,
	files []*domain.DriveFile,
	userId model.Id,
) (map[model.Id]FilePermissionResult, error) {
	result := make(map[model.Id]FilePermissionResult, len(files))
	if len(files) == 0 {
		return result, nil
	}

	// Phase 1 – owner check (in-memory, no DB)
	needShareCheck := make([]*domain.DriveFile, 0, len(files))
	for _, f := range files {
		if f == nil || f.Id == nil {
			continue
		}
		if f.OwnerRef != nil && *f.OwnerRef == userId {
			result[*f.Id] = filePermissionResult(f.IsFolder, enum.DriveFilePermOwner)
		} else {
			needShareCheck = append(needShareCheck, f)
		}
	}
	if len(needShareCheck) == 0 {
		return result, nil
	}

	// Phase 2 – batch direct shares (1 DB call)
	directFileIds := make([]model.Id, 0, len(needShareCheck))
	for _, f := range needShareCheck {
		directFileIds = append(directFileIds, *f.Id)
	}

	directShareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: directFileIds,
			UserId:       userId,
		},
	)
	if err != nil {
		return nil, err
	}

	directShareByFile := make(map[model.Id]enum.DriveFilePerm)
	if directShareResult != nil && directShareResult.ClientError == nil && directShareResult.HasData {
		for _, sh := range directShareResult.Data {
			if sh == nil {
				continue
			}
			if p := storedSharePerm(sh.Permission); p != enum.DriveFilePermNone {
				if existing, ok := directShareByFile[sh.FileRef]; !ok || p > existing {
					directShareByFile[sh.FileRef] = p
				}
			}
		}
	}

	needAncestorCheck := make([]*domain.DriveFile, 0)
	for _, f := range needShareCheck {
		if p, ok := directShareByFile[*f.Id]; ok {
			result[*f.Id] = filePermissionResult(f.IsFolder, p)
		} else {
			needAncestorCheck = append(needAncestorCheck, f)
		}
	}
	if len(needAncestorCheck) == 0 {
		return result, nil
	}

	// Phase 3 – batch ancestor IDs (1 DB call)
	ancestorFileIds := make([]model.Id, 0, len(needAncestorCheck))
	for _, f := range needAncestorCheck {
		ancestorFileIds = append(ancestorFileIds, *f.Id)
	}

	ancestorMap, err := this.driveFileRepo.GetAncestorIdsForFiles(ctx, ancestorFileIds)
	if err != nil {
		return nil, err
	}

	// Collect all unique ancestor IDs across all files
	allAncestorIdSet := collections.NewSet([]string{})
	for _, ids := range ancestorMap {
		for _, id := range ids {
			allAncestorIdSet.Add(string(id))
		}
	}
	allAncestorIds := make([]model.Id, 0, allAncestorIdSet.Len())
	for _, v := range allAncestorIdSet.GetValues() {
		allAncestorIds = append(allAncestorIds, model.Id(v))
	}

	if len(allAncestorIds) == 0 {
		for _, f := range needAncestorCheck {
			result[*f.Id] = filePermissionResult(f.IsFolder, enum.DriveFilePermNone)
		}
		return result, nil
	}

	// Phase 4a – batch load ancestor files to check ownership (1 DB call)
	ancestorFiles, err := this.driveFileRepo.FindByIds(ctx, allAncestorIds)
	if err != nil {
		return nil, err
	}

	ancestorOwnerById := make(map[model.Id]model.Id, len(ancestorFiles))
	for _, af := range ancestorFiles {
		if af == nil || af.Id == nil || af.OwnerRef == nil {
			continue
		}
		ancestorOwnerById[*af.Id] = *af.OwnerRef
	}

	// Phase 4b – batch ancestor shares (1 DB call)
	ancestorShareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: allAncestorIds,
			UserId:       userId,
		},
	)
	if err != nil {
		return nil, err
	}

	ancestorShareByFile := make(map[model.Id]enum.DriveFilePerm)
	if ancestorShareResult != nil && ancestorShareResult.ClientError == nil && ancestorShareResult.HasData {
		for _, sh := range ancestorShareResult.Data {
			if sh == nil {
				continue
			}
			if p := storedSharePerm(sh.Permission); p != enum.DriveFilePermNone {
				if existing, ok := ancestorShareByFile[sh.FileRef]; !ok || p > existing {
					ancestorShareByFile[sh.FileRef] = p
				}
			}
		}
	}

	// Phase 5 – resolve each remaining file from the pre-loaded data (in-memory)
	for _, f := range needAncestorCheck {
		fileId := *f.Id
		ancestorIds := ancestorMap[fileId]

		bestPerm := enum.DriveFilePermNone

		for _, ancId := range ancestorIds {
			if ownerRef, ok := ancestorOwnerById[ancId]; ok && ownerRef == userId {
				bestPerm = enum.DriveFilePermAncestorOwner
				break
			}
		}

		if bestPerm == enum.DriveFilePermNone {
			maxBase := enum.DriveFilePermNone
			for _, ancId := range ancestorIds {
				if p, ok := ancestorShareByFile[ancId]; ok && p > maxBase {
					maxBase = p
				}
			}
			if maxBase != enum.DriveFilePermNone {
				bestPerm = toInheritedPerm(maxBase)
			}
		}

		result[fileId] = filePermissionResult(f.IsFolder, bestPerm)
	}

	return result, nil
}

// enrichDriveFilesWithPermissions resolves permissions for a slice of files in
// batch and writes the result back into each DriveFile's ResolvedPermission field.
func (this *DriveFileServiceImpl) enrichDriveFilesWithPermissions(
	ctx crud.Context,
	files []*domain.DriveFile,
	userId model.Id,
) error {
	if len(files) == 0 {
		return nil
	}
	permMap, err := this.resolvePermissionsBatch(ctx, files, userId)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f == nil || f.Id == nil {
			continue
		}
		if p, ok := permMap[*f.Id]; ok {
			f.ResolvedPermission = &domain.DriveFileResolvedPermission{
				Permission: p.Permission,
			}
		}
	}
	return nil
}
