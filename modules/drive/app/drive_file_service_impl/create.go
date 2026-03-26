package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) CreateDriveFile(ctx crud.Context, cmd it.CreateDriveFileCommand) (result *it.CreateDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create drive file"); e != nil {
			err = e
		}
	}()
	tx, err := this.driveFileRepo.BeginTransaction(ctx)

	ctx.SetDbTranx(tx)

	result, err = crud.Create(ctx, crud.CreateParam[*domain.DriveFile, *it.CreateDriveFileCommand, it.CreateDriveFileResult]{
		Action:              "create drive file",
		Command:             &cmd,
		AssertBusinessRules: this.assertCreateDriveFileRules,
		RepoCreate:          this.driveFileRepo.Create,
		SetDefault:          this.setCreateDriveFileDefaults,
		Sanitize:            this.sanitizeDriveFile,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateDriveFileResult {
			return &it.CreateDriveFileResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.DriveFile) *it.CreateDriveFileResult {
			return &it.CreateDriveFileResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	if !result.HasData {
		tx.Rollback()
		return result, err
	}

	// Insert ancestor records for the new file
	if err := this.insertAncestorsForFile(ctx, *result.Data.Id, result.Data.ParentDriveFileRef); err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	// Recalculate size of parents
	if cmd.ParentDriveFileRef != nil {
		err := this.recalculateSizeOfParent(ctx, *cmd.ParentDriveFileRef, result.Data.Size, true)
		if err != nil {
			tx.Rollback()
			ft.PanicOnErr(err)
		}
	}

	tx.Commit()

	return result, err
}

func (this *DriveFileServiceImpl) setCreateDriveFileDefaults(d *domain.DriveFile) {
	d.SetDefaults()
	d.Status = enum.DriveFileStatusActive
	if !d.IsFolder {
		d.Storage = enum.DriveFileStorageS3
		d.BuildStorageInfo(this.config.GetStr(constants.S3StorageEndpoint))
	}

}

func (this *DriveFileServiceImpl) assertCreateDriveFileRules(ctx crud.Context, d *domain.DriveFile, vErrs *ft.ValidationErrors) error {
	var ownerRef model.Id
	if d.OwnerRef != nil {
		ownerRef = *d.OwnerRef
	}

	exists, err := this.driveFileRepo.ExistsByOwnerParentNameFolder(ctx, ownerRef, d.ParentDriveFileRef, d.Name, d.IsFolder)
	ft.PanicOnErr(err)

	if exists {
		vErrs.Append("name", "a file or folder with this name already exists in this location")
		return nil
	}

	if d.ParentDriveFileRef != nil {
		parent, err := this.driveFileRepo.FindById(ctx, *d.ParentDriveFileRef)
		ft.PanicOnErr(err)
		if parent == nil {
			vErrs.Append("parentDriveFileRef", "parent drive file not found")
			return nil
		}
		actorID := d.UserId
		if actorID == "" && d.OwnerRef != nil {
			actorID = *d.OwnerRef
		}
		if actorID != "" {
			if err := this.assertDriveFileActionAllowed(ctx, parent, actorID, func(p FilePermissionResult) bool {
				return p.CanCreateTo()
			}, vErrs); err != nil {
				return err
			}
			if vErrs.Count() > 0 {
				return nil
			}
		}
	}

	if d.IsFolder {
		return nil
	}

	if d.File == nil {
		vErrs.Append("file", "file is required when creating a file (not folder)")
		return nil
	}

	err = this.storageAdapter.UploadBucket(ctx.InnerContext(), file_storage.BucketDrive, d.StorageKey, d.File)
	if err != nil {
		this.logger.Error("create drive file: storage upload failed", err)
		ft.PanicOnErr(err)
	}

	return nil
}

