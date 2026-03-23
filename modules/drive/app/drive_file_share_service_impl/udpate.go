package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) UpdateDriveFileShare(ctx crud.Context, cmd it.UpdateDriveFileShareCommand) (
	*it.UpdateDriveFileShareResult, error) {
	return crud.Update(ctx, crud.UpdateParam[*domain.DriveFileShare, it.UpdateDriveFileShareCommand, it.UpdateDriveFileShareResult]{
		Action:       "update drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		RepoUpdate:   this.driveFileShareRepo.Update,
		Sanitize:     func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}
