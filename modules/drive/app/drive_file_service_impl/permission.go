package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	drive_file_share "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileServiceImpl) resolvePermission(
	ctx crud.Context,
	file *domain.DriveFile,
	userId model.Id,
) (enum.DriveFileSharePerm, error) {
	if file == nil || file.Id == nil {
		return enum.DriveFileSharePermNone, nil
	}

	if file.OwnerRef != nil && *file.OwnerRef == userId {
		return enum.DriveFileSharePermEditTrash, nil
	}

	directShareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: []model.Id{*file.Id},
			UserId:       userId,
		},
	)
	if err != nil {
		return enum.DriveFileSharePermNone, err
	}
	if directShareResult != nil && directShareResult.ClientError == nil && directShareResult.HasData {
		maxDirectPerm := enum.DriveFileSharePermNone
		for _, share := range directShareResult.Data {
			if share == nil {
				continue
			}
			if share.Permission > maxDirectPerm {
				maxDirectPerm = share.Permission
			}
		}
		if maxDirectPerm > enum.DriveFileSharePermNone {
			return maxDirectPerm, nil
		}
	}

	return this.resolvePermissionFromAncestors(ctx, *file.Id, userId)
}

func (this *DriveFileServiceImpl) resolvePermissionFromAncestors(
	ctx crud.Context,
	driveFileId model.Id,
	userId model.Id,
) (enum.DriveFileSharePerm, error) {
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, driveFileId)
	if err != nil {
		return enum.DriveFileSharePermNone, err
	}

	ancestorIds := make([]model.Id, 0, len(ancestors))
	for _, ancestor := range ancestors {
		if ancestor == nil || ancestor.Id == nil {
			continue
		}

		ancestorIds = append(ancestorIds, *ancestor.Id)
		if ancestor.OwnerRef != nil && *ancestor.OwnerRef == userId {
			return enum.DriveFileSharePermEditTrash, nil
		}
	}
	if len(ancestorIds) == 0 {
		return enum.DriveFileSharePermNone, nil
	}
	shareResult, err := this.driveFileShareService.ListDriveFileSharesByFileRefsAndUser(
		ctx,
		drive_file_share.ListDriveFileSharesByFileRefsAndUserQuery{
			DriveFileIds: ancestorIds,
			UserId:       userId,
		},
	)

	if err != nil {
		return enum.DriveFileSharePermNone, err
	}

	if shareResult != nil && shareResult.ClientError != nil {
		return enum.DriveFileSharePermNone, nil
	}

	if shareResult == nil || !shareResult.HasData || len(shareResult.Data) == 0 {
		return enum.DriveFileSharePermNone, nil
	}

	maxPerm := enum.DriveFileSharePermNone
	for _, share := range shareResult.Data {
		if share == nil {
			continue
		}
		if share.Permission > maxPerm {
			maxPerm = share.Permission
		}
	}

	return maxPerm, nil
}
