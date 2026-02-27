package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) UpdateDriveFileShare(ctx crud.Context, cmd it.UpdateDriveFileShareCommand) (
	*it.UpdateDriveFileShareResult, error) {
	if v := cmd.Validate(); v.Count() > 0 {
		return &it.UpdateDriveFileShareResult{ClientError: v.ToClientError()}, nil
	}
	return crud.Update(ctx, crud.UpdateParam[*domain.DriveFileShare, it.UpdateDriveFileShareCommand, it.UpdateDriveFileShareResult]{
		Action:       "update drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		AssertBusinessRules: func(c crud.Context, _ *domain.DriveFileShare, fromDb *domain.DriveFileShare, vErrs *ft.ValidationErrors) error {
			return this.assertUpdateDriveFileShareBusinessRules(c, cmd, fromDb, vErrs)
		},
		RepoUpdate: this.driveFileShareRepo.Update,
		Sanitize:   func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{ClientError: shareMutationForbiddenOrValidationClientError(vErrs)}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}

func (this *DriveFileShareServiceImpl) assertUpdateDriveFileShareBusinessRules(
	ctx crud.Context,
	cmd it.UpdateDriveFileShareCommand,
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
