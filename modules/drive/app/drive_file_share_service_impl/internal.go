package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/common/collections"
)

func (this *DriveFileShareServiceImpl) enrichDriveFileSharesWithUsers(
	ctx crud.Context,
	shares []*domain.DriveFileShare,
) error {
	if len(shares) == 0 {
		return nil
	}

	userIdSet := collections.NewSet([]string{})
	for _, s := range shares {
		if s == nil {
			continue
		}
		// model.Id is string alias, but drive_file_share.UserRef is a value (not pointer)
		userIdSet.Add(string(s.UserRef))
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

	for _, s := range shares {
		if s == nil {
			continue
		}
		if usersById == nil {
			s.User = nil
			continue
		}
		if u := usersById[s.UserRef]; u != nil {
			s.User = &domain.DriveFileShareUser{
				Id:          u.Id,
				DisplayName: u.DisplayName,
				Email:       u.Email,
				AvatarUrl:   u.AvatarUrl,
			}
		} else {
			s.User = nil
		}
	}

	return nil
}

func (this *DriveFileShareServiceImpl) assertDriveFileShareExists(ctx crud.Context, d *domain.DriveFileShare, vErrs *ft.ValidationErrors) (*domain.DriveFileShare, error) {
	if d == nil || d.Id == nil {
		return nil, nil
	}
	share, err := this.driveFileShareRepo.FindById(ctx, *d.Id)
	if err != nil {
		return nil, err
	}
	if share == nil {
		if vErrs != nil {
			vErrs.AppendNotFound("driveFileShareId", "drive file share")
		}
		return nil, nil
	}
	return share, nil
}
