package drive_file_service_impl

import (
	"time"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/collections"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) DeleteDriveFile(ctx crud.Context, cmd it.DeleteDriveFileCommand) (result *it.DeleteDriveFileResult, err error) {
	var tx *ent.Tx
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete drive file"); e != nil {
			err = e
			if tx != nil {
				_ = tx.Rollback()
			}
		}
	}()

	tx, err = this.driveFileRepo.BeginTransaction(ctx)
	ft.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	toFailureResultFunc := func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileResult {
		return &it.DeleteDriveFileResult{ClientError: vErrs.ToClientError()}
	}

	vErrs := util.ToPtr(fault.NewValidationErrors())
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return toFailureResultFunc(vErrs), nil
	}

	if driveFile.ParentDriveFileRef != nil {
		err = this.recalculateSizeOfParent(ctx, *driveFile.ParentDriveFileRef, driveFile.Size, false)
		ft.PanicOnErr(err)
	}

	if !driveFile.IsFolder {
		res, err := this.deleteExistsFile(ctx, driveFile, cmd)
		ft.PanicOnErr(err)
		ft.PanicOnErr(tx.Commit())

		return res, nil
	}

	err = this.deleteDriveFolder(ctx, driveFile)
	ft.PanicOnErr(err)
	ft.PanicOnErr(tx.Commit())

	deletedCount := 1

	return crud.NewSuccessDeletionResult(*driveFile.Id, &deletedCount), nil
}

func (this *DriveFileServiceImpl) deleteExistsFile(
	ctx crud.Context,
	driveFile *domain.DriveFile,
	cmd it.DeleteDriveFileCommand,
) (result *it.DeleteDriveFileResult, err error) {
	err = this.storageAdapter.DeleteBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.StorageKey)
	if err != nil {
		this.logger.Error("delete drive file: storage delete failed", err)
		ft.PanicOnErr(err)
	}

	result, err = crud.DeleteHard(ctx,
		crud.DeleteHardParam[*domain.DriveFile, it.DeleteDriveFileCommand, it.DeleteDriveFileResult]{
			Action:       "delete drive file",
			Command:      cmd,
			AssertExists: this.assertDriveFileExists,
			RepoDelete: func(ctx crud.Context, d *domain.DriveFile) (int, error) {
				return this.driveFileRepo.DeleteById(ctx, *d.Id)
			},
			ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileResult {
				return &it.DeleteDriveFileResult{ClientError: vErrs.ToClientError()}
			},
			ToSuccessResult: func(_ *domain.DriveFile, deletedCount int) *it.DeleteDriveFileResult {
				return crud.NewSuccessDeletionResult(cmd.DriveFileId, &deletedCount)
			},
		})
	ft.PanicOnErr(err)

	return result, nil
}

func (this *DriveFileServiceImpl) deleteDriveFolder(ctx crud.Context, driveFile *domain.DriveFile) error {
	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
	ft.PanicOnErr(err)
	driveFile.BuildTree(children)

	folderChildren := []*domain.DriveFile{}
	fileChildren := []*domain.DriveFile{}
	for _, child := range children {
		if child.IsFolder {
			folderChildren = append(folderChildren, child)
		} else {
			fileChildren = append(fileChildren, child)
		}
	}

	storageKeys := lo.Map(fileChildren, func(driveFile *domain.DriveFile, index int) string {
		return driveFile.StorageKey
	})

	deletedKeys, _, err := this.storageAdapter.DeleteBulk(ctx, storageKeys)
	if err != nil {
		this.logger.Error("[DriveFileService] this.storageAdapter.DeleteBulk error", err)
		ft.PanicOnErr(err)
	}

	deletedKeySet := collections.NewSet(deletedKeys)
	deletedDriveFileIds := make([]string, 0, len(deletedKeys)+len(folderChildren))
	for _, child := range children {
		if deletedKeySet.Has(child.StorageKey) {
			deletedDriveFileIds = append(deletedDriveFileIds, *child.Id)
		}
	}
	for _, child := range folderChildren {
		deletedDriveFileIds = append(deletedDriveFileIds, *child.Id)
	}

	deletedDriveFileIdSet := collections.NewSet(deletedDriveFileIds)
	failedDriveFileIdSet := collections.NewSet([]model.Id{})

	// post-order: folder chỉ được xóa nếu toàn bộ children xóa thành công
	var postOrderDelete func(driveFile *domain.DriveFile) bool
	postOrderDelete = func(driveFile *domain.DriveFile) bool {
		shouldDelete := true
		for _, child := range driveFile.Children {
			shouldDelete = shouldDelete && postOrderDelete(child)
		}
		shouldDelete = shouldDelete && deletedDriveFileIdSet.Has(*driveFile.Id)
		if !shouldDelete {
			deletedDriveFileIdSet.Remove(*driveFile.Id)
			failedDriveFileIdSet.Add(*driveFile.Id)
		}
		return shouldDelete
	}
	postOrderDelete(driveFile)

	_, err = this.driveFileRepo.DeleteByIds(ctx, deletedDriveFileIdSet.GetValues())
	ft.PanicOnErr(err)

	// update status pending-delete cho các file/folder xóa thất bại
	allFiles := append(children, driveFile)
	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, failedDriveFileIdSet.Len())
	for _, f := range allFiles {
		if failedDriveFileIdSet.Has(*f.Id) {
			updateCmds = append(updateCmds, it.UpdateDriveFileMetadataCommand{
				Id:     *f.Id,
				Etag:   *f.Etag,
				Status: enum.DriveFileStatusPendingDelete,
			})
		}
	}

	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
		DriveFiles: updateCmds,
	})
	ft.PanicOnErr(err)

	return nil
}

func (this *DriveFileServiceImpl) DeleteTrashedDriveFile(ctx crud.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete trashed drive file"); e != nil {
			err = e
		}
	}()
	threshold := time.Now().Add(-constants.TrashedFileRetentionPeriod)

	driveFiles, err := this.driveFileRepo.GetExpiredTrashedDriveFiles(ctx, threshold)
	ft.PanicOnErr(err)

	for _, f := range driveFiles {
		if f == nil || f.Id == nil {
			continue
		}
		_, err := this.DeleteDriveFile(ctx, it.DeleteDriveFileCommand{
			DriveFileId: *f.Id,
		})
		ft.PanicOnErr(err)
	}

	return nil
}
