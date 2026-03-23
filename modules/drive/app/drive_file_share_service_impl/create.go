package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) CreateDriveFileShare(ctx crud.Context, cmd it.CreateDriveFileShareCommand) (
	*it.CreateDriveFileShareResult, error) {
	return crud.Create(ctx, crud.CreateParam[*domain.DriveFileShare, *it.CreateDriveFileShareCommand, it.CreateDriveFileShareResult]{
		Action:              "create drive file share",
		Command:             &cmd,
		AssertBusinessRules: this.assertCreateDriveFileShareBussinessRules,
		RepoCreate:          this.driveFileShareRepo.Create,
		SetDefault:          this.setCreateDefaults,
		Sanitize:            func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}

// Create Bulk

func (this *DriveFileShareServiceImpl) CreateBulkDriveFileShares(ctx crud.Context, cmd it.CreateBulkDriveFileShareCommand) (
	*it.CreateBulkDriveFileShareResult, error) {
	return crud.CreateBulk(ctx, crud.CreateBulkParam[*domain.DriveFileShare, it.CreateBulkDriveFileShareCommand, it.CreateBulkDriveFileShareResult]{
		Action:         "create bulk drive file shares",
		Command:        cmd,
		RepoCreateBulk: this.repoCreateBulk,
		SetDefault:     this.setCreateDefaults,
		Sanitize:       func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateBulkDriveFileShareResult {
			return &it.CreateBulkDriveFileShareResult{ClientError: vErrs.ToClientError()}
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
		d.Permission = enum.DriveFileSharePermDefault
	}
}

func (this *DriveFileShareServiceImpl) assertCreateDriveFileShareBussinessRules(ctx crud.Context,
	d *domain.DriveFileShare, vErrs *ft.ValidationErrors) error {

	// assert file ref exists
	driveFile, err := this.driveFileRepo.FindById(ctx, d.FileRef)
	if err != nil {
		return err
	}

	if driveFile == nil {
		if vErrs != nil {
			vErrs.Append("driveFileId", "drive file not found")
		}
		return nil
	}

	// assert user ref exists
	exists, err, clientErr := this.identityCqrs.UserExists(ctx, d.UserRef)
	if err != nil {
		return err
	}

	if clientErr != nil {
		if vErrs != nil {
			vErrs.MergeClientError(clientErr)
		}

		return nil
	}

	if !exists {
		if vErrs != nil {
			vErrs.Append("userRef", "user not found")
		}

		return nil
	}

	return nil
}
