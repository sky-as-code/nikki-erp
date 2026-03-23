package drive_file_service_impl

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) RestoreDriveFile(ctx crud.Context, cmd it.RestoreDriveFileCommand) (result *it.RestoreDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "restore drive file"); e != nil {
			err = e
		}
	}()
	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	ft.PanicOnErr(err)
	ctx.SetDbTranx(tx)

	vErrs := util.ToPtr(fault.NewValidationErrors())
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), vErrs)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return &it.MoveDriveFileToTrashResult{ClientError: vErrs.ToClientError()}, nil
	}

	// recalculate parent size
	if cmd.ParentFileRef != nil && driveFile.ParentDriveFileRef != nil && *cmd.ParentFileRef != *driveFile.ParentDriveFileRef {
		if cmd.ParentFileRef != nil {
			err := this.recalculateSizeOfParent(ctx, *cmd.ParentFileRef, driveFile.Size, true)
			if err != nil {
				tx.Rollback()
				ft.PanicOnErr(err)
			}
		}

		if driveFile.ParentDriveFileRef != nil {
			err := this.recalculateSizeOfParent(ctx, *driveFile.ParentDriveFileRef, driveFile.Size, false)
			if err != nil {
				tx.Rollback()
				ft.PanicOnErr(err)
			}
		}
	}

	if driveFile.IsFolder {
		err := this.updateChildrenStatus(ctx, driveFile, enum.DriveFileStatusActive)
		if err != nil {
			tx.Rollback()
			ft.PanicOnErr(err)
		}
	}

	err = this.assertRestoreDriveFileRules(ctx, cmd.ToDomainModel(), driveFile, vErrs)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	if len(*vErrs) != 0 {
		tx.Rollback()
		return &it.RestoreDriveFileResult{ClientError: vErrs.ToClientError()}, nil
	}

	driveFile.Status = enum.DriveFileStatusActive
	driveFile.ParentDriveFileRef = cmd.ParentFileRef
	notDeleted := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	driveFile.DeletedAt = &notDeleted
	this.sanitizeDriveFile(driveFile)

	pathStr, err := this.buildMaterializedPathUnderParent(ctx, cmd.ParentFileRef, *driveFile.Id)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}
	driveFile.MaterializedPath = &pathStr

	driveFile, err = this.driveFileRepo.Overwrite(ctx, driveFile, *driveFile.Etag)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	if driveFile.IsFolder {
		if err := this.propagateMaterializedPathsAfterMove(ctx, driveFile, pathStr); err != nil {
			tx.Rollback()
			ft.PanicOnErr(err)
		}
	}

	ft.PanicOnErr(tx.Commit())
	return &it.RestoreDriveFileResult{HasData: true, Data: driveFile}, nil
}

func (this *DriveFileServiceImpl) assertRestoreDriveFileRules(
	ctx crud.Context,
	d *domain.DriveFile,
	fromDb *domain.DriveFile,
	vErrs *ft.ValidationErrors,
) error {
	if fromDb == nil {
		return nil
	}

	if fromDb.Status != enum.DriveFileStatusInTrash &&
		fromDb.Status != enum.DriveFileStatusParentInTrash {
		vErrs.Append("status", "only files in trash can be restored")
	}

	// check if not move to root
	if d.ParentDriveFileRef != nil {
		parent, err := this.driveFileRepo.FindById(ctx, *d.ParentDriveFileRef)
		ft.PanicOnErr(err)

		if parent == nil {
			vErrs.Append("parentFileRef", "parent drive file not found")
			return nil
		}

		if !parent.IsFolder {
			vErrs.Append("parentFileRef", "parent drive must be a folder")
			return nil
		}

		if parent.Status != enum.DriveFileStatusActive {
			vErrs.Append("parentFileRef", "parent drive must be active")
			return nil
		}
	}

	if vErrs.Count() > 0 {
		return nil
	}

	d.Status = enum.DriveFileStatusActive

	return nil
}
