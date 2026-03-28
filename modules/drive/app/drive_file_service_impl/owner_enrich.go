package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/collections"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

func (this *DriveFileServiceImpl) enrichDriveFilesWithOwners(ctx crud.Context, files []*domain.DriveFile) error {
	if len(files) == 0 {
		return nil
	}

	userIdSet := collections.NewSet([]string{})
	for _, f := range files {
		if f == nil || f.OwnerRef == nil || *f.OwnerRef == "" {
			continue
		}
		userIdSet.Add(string(*f.OwnerRef))
	}

	userIds := userIdSet.GetValues()
	if len(userIds) == 0 {
		return nil
	}

	ids := make([]model.Id, 0, len(userIds))
	for _, id := range userIds {
		ids = append(ids, model.Id(id))
	}

	usersById, err, _ := this.identityCqrs.GetUsersByIds(ctx, ids)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f == nil {
			continue
		}
		if f.OwnerRef == nil || *f.OwnerRef == "" {
			f.Owner = nil
			continue
		}
		if usersById == nil {
			f.Owner = nil
			continue
		}
		if u := usersById[*f.OwnerRef]; u != nil {
			f.Owner = &domain.DriveFileShareUser{
				Id:          u.Id,
				DisplayName: u.DisplayName,
				Email:       u.Email,
				AvatarUrl:   u.AvatarUrl,
			}
		} else {
			f.Owner = nil
		}
	}

	return nil
}
