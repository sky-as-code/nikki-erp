package drive_file_service_impl

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) MoveDriveFileToTrash(ctx crud.Context, cmd it.MoveDriveFileToTrashCommand) (result *it.MoveDriveFileToTrashResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "move drive file to trash"); e != nil {
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

	now := time.Now()
	driveFile.DeletedAt = &now
	driveFile.Status = enum.DriveFileStatusInTrash
	updated, err := this.driveFileRepo.Update(ctx, driveFile, *driveFile.Etag)
	ft.PanicOnErr(err)

	if driveFile.IsFolder {
		err = this.updateChildrenStatus(ctx, driveFile, enum.DriveFileStatusParentInTrash)
		if err != nil {
			tx.Rollback()
			ft.PanicOnErr(err)
		}
	}

	tx.Commit()

	return &it.MoveDriveFileToTrashResult{HasData: true, Data: updated}, nil
}
