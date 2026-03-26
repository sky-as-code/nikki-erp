package drive_file_share_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/collections"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
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

func driveFileShareFileFromDomain(f *domain.DriveFile) *domain.DriveFileShareFile {
	if f == nil || f.Id == nil {
		return nil
	}
	return &domain.DriveFileShareFile{
		Id:       *f.Id,
		Name:     f.Name,
		IsFolder: f.IsFolder,
	}
}

func (this *DriveFileShareServiceImpl) enrichDriveFileSharesWithFiles(
	ctx crud.Context,
	shares []*domain.DriveFileShare,
) error {
	if len(shares) == 0 {
		return nil
	}

	fileIdSet := collections.NewSet([]string{})
	for _, s := range shares {
		if s == nil || s.FileRef == "" {
			continue
		}
		fileIdSet.Add(string(s.FileRef))
	}
	ids := fileIdSet.GetValues()
	if len(ids) == 0 {
		return nil
	}

	filesById := make(map[model.Id]*domain.DriveFile, len(ids))
	for _, idStr := range ids {
		f, err := this.driveFileRepo.FindById(ctx, model.Id(idStr))
		if err != nil {
			return err
		}
		if f != nil && f.Id != nil {
			filesById[*f.Id] = f
		}
	}

	for _, s := range shares {
		if s == nil {
			continue
		}
		if f := filesById[s.FileRef]; f != nil {
			s.File = driveFileShareFileFromDomain(f)
		} else {
			s.File = nil
		}
	}

	return nil
}

func (this *DriveFileShareServiceImpl) enrichDriveFileSharesWithViews(ctx crud.Context, shares []*domain.DriveFileShare) error {
	if err := this.enrichDriveFileSharesWithUsers(ctx, shares); err != nil {
		return err
	}
	return this.enrichDriveFileSharesWithFiles(ctx, shares)
}

func (this *DriveFileShareServiceImpl) enrichDriveFileUserShareDetailsWithFiles(ctx crud.Context, out []*it.DriveFileUserShareDetail) error {
	if len(out) == 0 {
		return nil
	}

	fileIdSet := collections.NewSet([]string{})
	for _, item := range out {
		if item == nil || item.FileRef == "" {
			continue
		}
		fileIdSet.Add(string(item.FileRef))
	}
	ids := fileIdSet.GetValues()
	if len(ids) == 0 {
		return nil
	}

	filesById := make(map[model.Id]*domain.DriveFile, len(ids))
	for _, idStr := range ids {
		f, err := this.driveFileRepo.FindById(ctx, model.Id(idStr))
		if err != nil {
			return err
		}
		if f != nil && f.Id != nil {
			filesById[*f.Id] = f
		}
	}

	for _, item := range out {
		if item == nil {
			continue
		}
		if f := filesById[item.FileRef]; f != nil {
			item.File = driveFileShareFileFromDomain(f)
		} else {
			item.File = nil
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
