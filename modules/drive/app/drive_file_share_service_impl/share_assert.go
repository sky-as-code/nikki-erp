package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

const validationKeyForbidden = "forbidden"

// shareMutationForbiddenOrValidationClientError maps validation errors: reserved key "forbidden" -> 403 client error.
func shareMutationForbiddenOrValidationClientError(vErrs *ft.ValidationErrors) *ft.ClientError {
	if vErrs == nil || vErrs.Count() == 0 {
		return nil
	}
	if msg, ok := (*vErrs)[validationKeyForbidden]; ok {
		return &ft.ClientError{Code: "forbidden", Details: msg}
	}
	return vErrs.ToClientError()
}

// assertActorIsOwnerOrAncestorOwnerOfDriveFile requires driveFile non-nil with Id set.
func (this *DriveFileShareServiceImpl) assertActorIsOwnerOrAncestorOwnerOfDriveFile(
	ctx crud.Context,
	driveFile *domain.DriveFile,
	actorId model.Id,
	vErrs *ft.ValidationErrors,
) error {
	if driveFile == nil || driveFile.Id == nil {
		return nil
	}
	perm, err := this.permissionSvc.ResolvePermission(ctx, driveFile, actorId)
	if err != nil {
		return err
	}
	if perm.Permission == enum.DriveFilePermOwner || perm.Permission == enum.DriveFilePermAncestorOwner {
		return nil
	}
	if vErrs != nil {
		vErrs.Append(validationKeyForbidden, "only the file owner or an ancestor folder owner can manage shares")
	}
	return nil
}

// assertShareActorMayManageFile loads the drive file, asserts it exists, then owner/ancestor-owner for actorId.
func (this *DriveFileShareServiceImpl) assertShareActorMayManageFile(
	ctx crud.Context,
	fileRef model.Id,
	actorId model.Id,
	vErrs *ft.ValidationErrors,
) error {
	driveFile, err := this.driveFileRepo.FindById(ctx, fileRef)
	if err != nil {
		return err
	}
	if driveFile == nil {
		if vErrs != nil {
			vErrs.Append("driveFileId", "drive file not found")
		}
		return nil
	}
	return this.assertActorIsOwnerOrAncestorOwnerOfDriveFile(ctx, driveFile, actorId, vErrs)
}

func (this *DriveFileShareServiceImpl) assertShareTargetUserExists(
	ctx crud.Context,
	userRef model.Id,
	vErrs *ft.ValidationErrors,
) error {
	exists, err, clientErr := this.identityCqrs.UserExists(ctx, userRef)
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
	}
	return nil
}
