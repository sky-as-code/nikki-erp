package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) DeleteDriveFileShare(ctx crud.Context, cmd it.DeleteDriveFileShareCommand) (
	*it.DeleteDriveFileShareResult, error) {
	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.DriveFileShare, it.DeleteDriveFileShareCommand, it.DeleteDriveFileShareResult]{
		Action:       "delete drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		RepoDelete: func(ctx crud.Context, d *domain.DriveFileShare) (int, error) {
			return this.driveFileShareRepo.DeleteById(ctx, *d.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileShareResult {
			return &it.DeleteDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(_ *domain.DriveFileShare, deletedCount int,
		) *it.DeleteDriveFileShareResult {
			return crud.NewSuccessDeletionResult(cmd.DriveFileShareId, &deletedCount)
		},
	})
}
