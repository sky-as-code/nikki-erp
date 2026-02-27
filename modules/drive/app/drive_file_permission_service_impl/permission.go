package drive_file_permission_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/collections"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFilePermissionServiceImpl) listSharesByFileRefsAndUser(
	ctx crud.Context,
	driveFileIds []model.Id,
	userId model.Id,
) ([]*domain.DriveFileShare, error) {
	if len(driveFileIds) == 0 {
		return nil, nil
	}
	return this.driveFileShareRepo.ListByFileRefsAndUserRef(ctx, driveFileIds, userId)
}

func filePermissionResult(isFolder bool, perm enum.DriveFilePerm) it.FilePermissionResult {
	return it.FilePermissionResult{IsFolder: isFolder, Permission: perm}
}

func storedSharePerm(p enum.DriveFilePerm) enum.DriveFilePerm {
	switch p {
	case enum.DriveFilePermView, enum.DriveFilePermEdit, enum.DriveFilePermEditTrash:
		return p
	default:
		return enum.DriveFilePermNone
	}
}

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

func (this *DriveFilePermissionServiceImpl) ResolvePermission(
	ctx crud.Context,
	file *domain.DriveFile,
	userId model.Id,
) (it.FilePermissionResult, error) {
	if file == nil || file.Id == nil {
		return filePermissionResult(false, enum.DriveFilePermNone), nil
	}

	isFolder := file.IsFolder

	if file.OwnerRef != nil && *file.OwnerRef == userId {
		return filePermissionResult(isFolder, enum.DriveFilePermOwner), nil
	}

	directShares, err := this.listSharesByFileRefsAndUser(ctx, []model.Id{*file.Id}, userId)
	if err != nil {
		return filePermissionResult(isFolder, enum.DriveFilePermNone), err
	}
	directPerm := enum.DriveFilePermNone
	for _, sh := range directShares {
		if sh == nil {
			continue
		}
		if p := storedSharePerm(sh.Permission); p > directPerm {
			directPerm = p
		}
	}

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

func (this *DriveFilePermissionServiceImpl) resolvePermissionFromAncestors(
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
	ancestorShares, err := this.listSharesByFileRefsAndUser(ctx, ancestorIds, userId)
	if err != nil {
		return enum.DriveFilePermNone, err
	}

	if len(ancestorShares) == 0 {
		return enum.DriveFilePermNone, nil
	}

	maxBase := maxStoredSharePermAcrossAncestors(ancestorShares)
	if maxBase == enum.DriveFilePermNone {
		return enum.DriveFilePermNone, nil
	}

	return toInheritedPerm(maxBase), nil
}

// maxStoredSharePermByFileRef returns the strongest stored share level per file_ref.
func maxStoredSharePermByFileRef(shares []*domain.DriveFileShare) map[model.Id]enum.DriveFilePerm {
	out := make(map[model.Id]enum.DriveFilePerm)
	for _, sh := range shares {
		if sh == nil {
			continue
		}
		if p := storedSharePerm(sh.Permission); p != enum.DriveFilePermNone {
			if existing, ok := out[sh.FileRef]; !ok || p > existing {
				out[sh.FileRef] = p
			}
		}
	}
	return out
}

func partitionBatchOwners(
	files []*domain.DriveFile,
	userId model.Id,
) (result map[model.Id]it.FilePermissionResult, needShareCheck []*domain.DriveFile) {
	result = make(map[model.Id]it.FilePermissionResult, len(files))
	needShareCheck = make([]*domain.DriveFile, 0, len(files))
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
	return result, needShareCheck
}

func uniqueAncestorIdsFromMap(ancestorMap map[model.Id][]model.Id) []model.Id {
	set := collections.NewSet([]string{})
	for _, ids := range ancestorMap {
		for _, id := range ids {
			set.Add(string(id))
		}
	}
	out := make([]model.Id, 0, set.Len())
	for _, v := range set.GetValues() {
		out = append(out, model.Id(v))
	}
	return out
}

func resolvePermFromAncestorChain(
	isFolder bool,
	ancestorIds []model.Id,
	userId model.Id,
	ancestorOwnerById map[model.Id]model.Id,
	ancestorShareByFile map[model.Id]enum.DriveFilePerm,
) it.FilePermissionResult {
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
	return filePermissionResult(isFolder, bestPerm)
}

func (this *DriveFilePermissionServiceImpl) batchApplyDirectShares(
	ctx crud.Context,
	needShareCheck []*domain.DriveFile,
	userId model.Id,
	result map[model.Id]it.FilePermissionResult,
) ([]*domain.DriveFile, error) {
	directFileIds := make([]model.Id, 0, len(needShareCheck))
	for _, f := range needShareCheck {
		directFileIds = append(directFileIds, *f.Id)
	}
	directShares, err := this.listSharesByFileRefsAndUser(ctx, directFileIds, userId)
	if err != nil {
		return nil, err
	}
	directShareByFile := maxStoredSharePermByFileRef(directShares)

	needAncestorCheck := make([]*domain.DriveFile, 0)
	for _, f := range needShareCheck {
		if p, ok := directShareByFile[*f.Id]; ok {
			result[*f.Id] = filePermissionResult(f.IsFolder, p)
		} else {
			needAncestorCheck = append(needAncestorCheck, f)
		}
	}
	return needAncestorCheck, nil
}

func (this *DriveFilePermissionServiceImpl) loadAncestorOwnerAndShareMaps(
	ctx crud.Context,
	allAncestorIds []model.Id,
	userId model.Id,
) (ancestorOwnerById map[model.Id]model.Id, ancestorShareByFile map[model.Id]enum.DriveFilePerm, err error) {
	ancestorFiles, err := this.driveFileRepo.FindByIds(ctx, allAncestorIds)
	if err != nil {
		return nil, nil, err
	}
	ancestorOwnerById = make(map[model.Id]model.Id, len(ancestorFiles))
	for _, af := range ancestorFiles {
		if af == nil || af.Id == nil || af.OwnerRef == nil {
			continue
		}
		ancestorOwnerById[*af.Id] = *af.OwnerRef
	}
	ancestorShares, err := this.listSharesByFileRefsAndUser(ctx, allAncestorIds, userId)
	if err != nil {
		return nil, nil, err
	}
	ancestorShareByFile = maxStoredSharePermByFileRef(ancestorShares)
	return ancestorOwnerById, ancestorShareByFile, nil
}

func (this *DriveFilePermissionServiceImpl) ResolvePermissionsBatch(
	ctx crud.Context,
	files []*domain.DriveFile,
	userId model.Id,
) (map[model.Id]it.FilePermissionResult, error) {
	if len(files) == 0 {
		return make(map[model.Id]it.FilePermissionResult), nil
	}

	result, needShareCheck := partitionBatchOwners(files, userId)
	if len(needShareCheck) == 0 {
		return result, nil
	}

	needAncestorCheck, err := this.batchApplyDirectShares(ctx, needShareCheck, userId, result)
	if err != nil {
		return nil, err
	}
	if len(needAncestorCheck) == 0 {
		return result, nil
	}

	ancestorFileIds := make([]model.Id, 0, len(needAncestorCheck))
	for _, f := range needAncestorCheck {
		ancestorFileIds = append(ancestorFileIds, *f.Id)
	}
	ancestorMap, err := this.driveFileRepo.GetAncestorIdsForFiles(ctx, ancestorFileIds)
	if err != nil {
		return nil, err
	}

	allAncestorIds := uniqueAncestorIdsFromMap(ancestorMap)
	if len(allAncestorIds) == 0 {
		for _, f := range needAncestorCheck {
			result[*f.Id] = filePermissionResult(f.IsFolder, enum.DriveFilePermNone)
		}
		return result, nil
	}

	ancestorOwnerById, ancestorShareByFile, err := this.loadAncestorOwnerAndShareMaps(ctx, allAncestorIds, userId)
	if err != nil {
		return nil, err
	}

	for _, f := range needAncestorCheck {
		fileId := *f.Id
		result[fileId] = resolvePermFromAncestorChain(
			f.IsFolder,
			ancestorMap[fileId],
			userId,
			ancestorOwnerById,
			ancestorShareByFile,
		)
	}

	return result, nil
}

func (this *DriveFilePermissionServiceImpl) EnrichDriveFilesWithPermissions(
	ctx crud.Context,
	files []*domain.DriveFile,
	userId model.Id,
) error {
	if len(files) == 0 {
		return nil
	}
	permMap, err := this.ResolvePermissionsBatch(ctx, files, userId)
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

func (this *DriveFilePermissionServiceImpl) AssertDriveFileActionAllowed(
	ctx crud.Context,
	file *domain.DriveFile,
	userId model.Id,
	allow func(it.FilePermissionResult) bool,
	vErrs *ft.ValidationErrors,
) error {
	if file == nil || file.Id == nil || userId == "" {
		return nil
	}
	perm, err := this.ResolvePermission(ctx, file, userId)
	if err != nil {
		return err
	}
	if !allow(perm) {
		vErrs.AppendNotAllowed("driveFileId", "drive file")
	}
	return nil
}
