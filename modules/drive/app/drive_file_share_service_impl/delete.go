package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) DeleteDriveFileShare(ctx crud.Context, cmd it.DeleteDriveFileShareCommand) (
	*it.DeleteDriveFileShareResult, error) {
	if v := cmd.Validate(); v.Count() > 0 {
		return &it.DeleteDriveFileShareResult{ClientError: v.ToClientError()}, nil
	}
	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.DriveFileShare, it.DeleteDriveFileShareCommand, it.DeleteDriveFileShareResult]{
		Action:       "delete drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		AssertBusinessRules: func(c crud.Context, command it.DeleteDriveFileShareCommand, modelFromDb *domain.DriveFileShare, vErrs *ft.ValidationErrors) error {
			return this.assertDeleteDriveFileShareBusinessRules(c, command, modelFromDb, vErrs)
		},
		RepoDelete: func(ctx crud.Context, d *domain.DriveFileShare) (int, error) {
			return this.driveFileShareRepo.DeleteById(ctx, *d.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileShareResult {
			return &it.DeleteDriveFileShareResult{ClientError: shareMutationForbiddenOrValidationClientError(vErrs)}
		},
		ToSuccessResult: func(_ *domain.DriveFileShare, deletedCount int,
		) *it.DeleteDriveFileShareResult {
			return crud.NewSuccessDeletionResult(cmd.DriveFileShareId, &deletedCount)
		},
	})
}

func (this *DriveFileShareServiceImpl) assertDeleteDriveFileShareBusinessRules(
	ctx crud.Context,
	cmd it.DeleteDriveFileShareCommand,
	shareFromDb *domain.DriveFileShare,
	vErrs *ft.ValidationErrors,
) error {
	if shareFromDb == nil {
		return nil
	}
	if shareFromDb.FileRef != cmd.DriveFileId {
		vErrs.AppendNotAllowed("driveFileShareId", "drive file share")
		return nil
	}
	driveFile, err := this.driveFileRepo.FindById(ctx, shareFromDb.FileRef)
	if err != nil {
		return err
	}
	if driveFile == nil {
		vErrs.Append("driveFileId", "drive file not found")
		return nil
	}
	return this.assertActorIsOwnerOrAncestorOwnerOfDriveFile(ctx, driveFile, cmd.UserId, vErrs)
}
