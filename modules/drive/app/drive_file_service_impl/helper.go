package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) recalculateSizeOfParent(ctx crud.Context,
	parentId model.Id, sizeDelta uint64, inc bool) error {
	driveFiles, err := this.driveFileRepo.GetDriveFileParents(ctx, parentId)
	ft.PanicOnErr(err)

	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, len(driveFiles))
	for _, f := range driveFiles {
		cmd := it.UpdateDriveFileMetadataCommand{
			Id:   *f.Id,
			Etag: *f.Etag,
		}

		if inc {
			cmd.Size = f.Size + uint64(sizeDelta)
		} else {
			cmd.Size = f.Size - uint64(sizeDelta)
		}

		updateCmds = append(updateCmds, cmd)
	}

	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
		DriveFiles: updateCmds,
	})
	ft.PanicOnErr(err)

	return nil
}

func (this *DriveFileServiceImpl) updateChildrenStatus(ctx crud.Context, driveFile *domain.DriveFile, status enum.DriveFileStatus) error {
	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
	ft.PanicOnErr(err)

	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, len(children))
	for _, f := range children {
		updateCmds = append(updateCmds, it.UpdateDriveFileMetadataCommand{
			Id:     *f.Id,
			Etag:   *f.Etag,
			Status: status,
		})
	}

	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
		DriveFiles: updateCmds,
	})
	ft.PanicOnErr(err)

	return nil
}

func (this *DriveFileServiceImpl) sanitizeDriveFile(d *domain.DriveFile) {
	if d != nil && d.Name != "" {
		d.Name = defense.SanitizePlainText(d.Name, true)
	}
}

func (this *DriveFileServiceImpl) assertDriveFileActionAllowed(
	ctx crud.Context,
	file *domain.DriveFile,
	userId model.Id,
	allow func(FilePermissionResult) bool,
	vErrs *ft.ValidationErrors,
) error {
	if file == nil || file.Id == nil || userId == "" {
		return nil
	}
	perm, err := this.resolvePermission(ctx, file, userId)
	if err != nil {
		return err
	}
	if !allow(perm) {
		vErrs.AppendNotAllowed("driveFileId", "drive file")
	}
	return nil
}

func (this *DriveFileServiceImpl) assertDriveFileExists(ctx crud.Context, d *domain.DriveFile, vErrs *ft.ValidationErrors) (*domain.DriveFile, error) {
	if d == nil || d.Id == nil {
		return nil, nil
	}

	driveFile, err := this.driveFileRepo.FindById(ctx, *d.Id)
	ft.PanicOnErr(err)

	if driveFile == nil {
		if vErrs != nil {
			vErrs.Append("driveFileId", "drive file not found")
		}
		return nil, nil
	}

	return driveFile, nil
}
