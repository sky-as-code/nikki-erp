package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) CreateDriveFileShare(ctx crud.Context, cmd it.CreateDriveFileShareCommand) (
	*it.CreateDriveFileShareResult, error) {
	if v := cmd.Validate(); v.Count() > 0 {
		return &it.CreateDriveFileShareResult{ClientError: v.ToClientError()}, nil
	}
	return crud.Create(ctx, crud.CreateParam[*domain.DriveFileShare, *it.CreateDriveFileShareCommand, it.CreateDriveFileShareResult]{
		Action:  "create drive file share",
		Command: &cmd,
		AssertBusinessRules: func(c crud.Context, d *domain.DriveFileShare, vErrs *ft.ValidationErrors) error {
			return this.assertCreateDriveFileShareBussinessRules(c, d, cmd.UserId, vErrs)
		},
		RepoCreate: this.driveFileShareRepo.Create,
		SetDefault: this.setCreateDefaults,
		Sanitize:   func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{ClientError: shareMutationForbiddenOrValidationClientError(vErrs)}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}

func (this *DriveFileShareServiceImpl) CreateBulkDriveFileShares(ctx crud.Context, cmd it.CreateBulkDriveFileShareCommand) (
	*it.CreateBulkDriveFileShareResult, error) {
	if v := cmd.Validate(); v.Count() > 0 {
		return &it.CreateBulkDriveFileShareResult{ClientError: v.ToClientError()}, nil
	}
	vErrsPre := ft.NewValidationErrors()
	if err := this.assertShareActorMayManageFile(ctx, cmd.FileRef, cmd.UserId, &vErrsPre); err != nil {
		return nil, err
	}
	if vErrsPre.Count() > 0 {
		return &it.CreateBulkDriveFileShareResult{
			ClientError: shareMutationForbiddenOrValidationClientError(&vErrsPre),
		}, nil
	}

	return crud.CreateBulk(ctx, crud.CreateBulkParam[*domain.DriveFileShare, it.CreateBulkDriveFileShareCommand, it.CreateBulkDriveFileShareResult]{
		Action:              "create bulk drive file shares",
		Command:             cmd,
		AssertBusinessRules: this.assertShareTargetUserExistsForCreateBulk,
		RepoCreateBulk:      this.repoCreateBulk,
		SetDefault:          this.setCreateDefaults,
		Sanitize:            func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateBulkDriveFileShareResult {
			return &it.CreateBulkDriveFileShareResult{ClientError: shareMutationForbiddenOrValidationClientError(vErrs)}
		},
		ToSuccessResult: func(models []*domain.DriveFileShare) *it.CreateBulkDriveFileShareResult {
			return &it.CreateBulkDriveFileShareResult{HasData: models != nil, Data: models}
		},
	})
}

func (this *DriveFileShareServiceImpl) repoCreateBulk(ctx crud.Context, models []*domain.DriveFileShare) ([]*domain.DriveFileShare, error) {
	created := make([]*domain.DriveFileShare, 0, len(models))
	for _, m := range models {
		result, err := this.driveFileShareRepo.Create(ctx, m)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}
	return created, nil
}

func (this *DriveFileShareServiceImpl) setCreateDefaults(d *domain.DriveFileShare) {
	d.SetDefaults()
	if d.Permission == 0 {
		d.Permission = enum.DriveFilePermDefault
	}
}

func (this *DriveFileShareServiceImpl) assertShareTargetUserExistsForCreateBulk(
	ctx crud.Context,
	d *domain.DriveFileShare,
	vErrs *ft.ValidationErrors,
) error {
	return this.assertShareTargetUserExists(ctx, d.UserRef, vErrs)
}

func (this *DriveFileShareServiceImpl) assertCreateDriveFileShareBussinessRules(ctx crud.Context,
	d *domain.DriveFileShare, actorId model.Id, vErrs *ft.ValidationErrors) error {

	if err := this.assertShareActorMayManageFile(ctx, d.FileRef, actorId, vErrs); err != nil {
		return err
	}
	if vErrs.Count() > 0 {
		return nil
	}

	return this.assertShareTargetUserExists(ctx, d.UserRef, vErrs)
}
