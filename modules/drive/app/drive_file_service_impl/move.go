package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) MoveDriveFile(ctx crud.Context, cmd it.MoveDriveFileCommand) (result *it.MoveDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "move drive file"); e != nil {
			err = e
		}
	}()
	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	ft.PanicOnErr(err)
	ctx.SetDbTranx(tx)

	vErrs := ft.NewValidationErrors()
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), &vErrs)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return &it.MoveDriveFileResult{ClientError: vErrs.ToClientError()}, nil
	}

	if driveFile.IsFolder {
		err := this.updateChildrenStatus(ctx, driveFile, enum.DriveFileStatusActive)
		if err != nil {
			tx.Rollback()
			ft.PanicOnErr(err)
		}
	}

	err = this.assertMoveDriveFileRules(ctx, cmd.ToDomainModel(), driveFile, &vErrs)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	if len(vErrs) != 0 {
		return &it.MoveDriveFileResult{ClientError: vErrs.ToClientError()}, nil
	}

	driveFile.ParentDriveFileRef = cmd.ParentFileRef
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
	return &it.MoveDriveFileResult{HasData: true, Data: driveFile}, nil
}

func (this *DriveFileServiceImpl) assertMoveDriveFileRules(
	ctx crud.Context,
	d *domain.DriveFile,
	fromDb *domain.DriveFile,
	vErrs *ft.ValidationErrors,
) error {
	if fromDb == nil {
		return nil
	}

	if d.ParentDriveFileRef != nil {
		parent, err := this.driveFileRepo.FindById(ctx, *d.ParentDriveFileRef)
		ft.PanicOnErr(err)

		if parent == nil {
			vErrs.Append("parentFileRef", "parent drive file not found")
		}

		if !parent.IsFolder {
			vErrs.Append("parentFileRef", "parent drive must be a folder")
		}

		if parent.Status != enum.DriveFileStatusActive {
			vErrs.Append("parentFileRef", "parent drive must be active")
		}

		if parent.OwnerRef != d.OwnerRef {
			vErrs.Append("parentFileRef", "parent drive must be owned by the same user")
		}
	}

	return nil
}
