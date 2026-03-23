package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) UpdateDriveFileMetadata(ctx crud.Context, cmd it.UpdateDriveFileMetadataCommand) (result *it.UpdateDriveFileMetadataResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update drive file metadata"); e != nil {
			err = e
		}
	}()

	result, err = crud.Update(ctx, crud.UpdateParam[*domain.DriveFile, it.UpdateDriveFileMetadataCommand, it.UpdateDriveFileMetadataResult]{
		Action:       "update drive file metadata",
		Command:      cmd,
		AssertExists: this.assertDriveFileExists,
		RepoUpdate:   this.driveFileRepo.Update,
		Sanitize:     this.sanitizeDriveFile,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateDriveFileMetadataResult {
			return &it.UpdateDriveFileMetadataResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFile) *it.UpdateDriveFileMetadataResult {
			return &it.UpdateDriveFileMetadataResult{HasData: true, Data: d}
		},
	})
	return result, err
}

// Update bulk metadata

func (this *DriveFileServiceImpl) UpdateBulkDriveFileMetadata(ctx crud.Context, cmd it.UpdateBulkDriveFileMetadataCommand) (result *it.UpdateBulkDriveFileMetadataResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update bulk drive file metadata"); e != nil {
			err = e
		}
	}()
	result, err = crud.UpdateBulk(ctx, crud.UpdateBulkParam[*domain.DriveFile, it.UpdateBulkDriveFileMetadataCommand, it.UpdateBulkDriveFileMetadataResult]{
		Action:       "update bulk drive file metadata",
		Command:      cmd,
		AssertExists: this.assertDriveFileExists,
		RepoUpdate:   this.driveFileRepo.Update,
		Sanitize:     this.sanitizeDriveFile,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateBulkDriveFileMetadataResult {
			return &it.UpdateBulkDriveFileMetadataResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(models []*domain.DriveFile) *it.UpdateBulkDriveFileMetadataResult {
			return &it.UpdateBulkDriveFileMetadataResult{HasData: models != nil, Data: models}
		},
	})
	return result, err
}

// Update content

func (this *DriveFileServiceImpl) UpdateDriveFileContent(ctx crud.Context, cmd it.UpdateDriveFileContentCommand) (result *it.UpdateDriveFileContentResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update drive file content"); e != nil {
			err = e
		}
	}()
	result, err = crud.Update(ctx, crud.UpdateParam[*domain.DriveFile, it.UpdateDriveFileContentCommand, it.UpdateDriveFileContentResult]{
		Action:              "update drive file content",
		Command:             cmd,
		AssertExists:        this.assertDriveFileExists,
		AssertBusinessRules: this.assertUpdateDriveFileContentRules,
		RepoUpdate:          this.driveFileRepo.Update,
		Sanitize:            this.sanitizeDriveFile,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateDriveFileContentResult {
			return &it.UpdateDriveFileContentResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFile) *it.UpdateDriveFileContentResult {
			return &it.UpdateDriveFileContentResult{HasData: true, Data: d}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) assertUpdateDriveFileContentRules(ctx crud.Context, d *domain.DriveFile, fromDb *domain.DriveFile, vErrs *ft.ValidationErrors) error {
	if d.File == nil {
		vErrs.Append("file", "file is required when updating content")
		return nil
	}

	if fromDb == nil {
		return nil
	}

	if fromDb.IsFolder {
		vErrs.Append("driveFileId", "cannot update content of a folder")
		return nil
	}

	d.OwnerRef = fromDb.OwnerRef
	d.BuildStorageInfo(this.config.GetStr(constants.S3StorageEndpoint))

	if err := this.storageAdapter.DeleteBucket(ctx.InnerContext(), file_storage.BucketDrive, fromDb.StorageKey); err != nil {
		this.logger.Error("update drive file content: storage delete failed", err)
		ft.PanicOnErr(err)
	}

	if err := this.storageAdapter.UploadBucket(ctx.InnerContext(), file_storage.BucketDrive, d.StorageKey, d.File); err != nil {
		this.logger.Error("update drive file content: storage upload failed", err)
		ft.PanicOnErr(err)
	}

	return nil
}
