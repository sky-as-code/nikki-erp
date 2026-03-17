package app

import (
	"io"
	"time"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/collections"
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileServiceImpl struct {
	logger         logging.LoggerService
	config         config.ConfigService
	driveFileRepo  it.DriveFileRepository
	storageAdapter file_storage.FileStorageAdapter
}

func NewDriveFileService(
	logger logging.LoggerService,
	config config.ConfigService,
	driveFileRepo it.DriveFileRepository,
	storageAdapter file_storage.FileStorageAdapter,
) it.DriveFileService {
	return &DriveFileServiceImpl{
		logger:         logger,
		config:         config,
		driveFileRepo:  driveFileRepo,
		storageAdapter: storageAdapter,
	}
}

// Create
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

// Update metadata

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

// Delete

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

	toFailureResultFunc := func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileResult {
		return &it.DeleteDriveFileResult{ClientError: vErrs.ToClientError()}
	}

	vErrs := util.ToPtr(fault.NewValidationErrors())
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), vErrs)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return toFailureResultFunc(vErrs), nil
	}

	if !driveFile.IsFolder {
		return this.deleteExistsFile(ctx, driveFile, cmd)
	}

	tx, err = this.driveFileRepo.BeginTransaction(ctx)
	ft.PanicOnErr(err)
	ctx.SetDbTranx(tx)

	err = this.deleteDriveFolder(ctx, driveFile)
	ft.PanicOnErr(err)

	ft.PanicOnErr(tx.Commit())
	return nil, nil
}

func (this *DriveFileServiceImpl) deleteExistsFile(ctx crud.Context, driveFile *domain.DriveFile, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	err := this.storageAdapter.DeleteBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.StorageKey)
	if err != nil {
		this.logger.Error("delete drive file: storage delete failed", err)
		ft.PanicOnErr(err)
	}

	result, err := crud.DeleteHard(ctx,
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
	this.logger.Error("[DriveFileService] this.storageAdapter.DeleteBulk error", err)

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

// Get by ID

func (this *DriveFileServiceImpl) GetDriveFileById(ctx crud.Context, query it.GetDriveFileByIdQuery) (result *it.GetDriveFileByIdResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file by id"); e != nil {
			err = e
		}
	}()
	result, err = crud.GetOne(ctx, crud.GetOneParam[*domain.DriveFile, it.GetDriveFileByIdQuery, it.GetDriveFileByIdResult]{
		Action: "get drive file by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetDriveFileByIdQuery, vErrs *ft.ValidationErrors) (*domain.DriveFile, error) {
			driveFile, err := this.driveFileRepo.FindById(ctx, q.DriveFileId)
			ft.PanicOnErr(err)
			if driveFile == nil {
				vErrs.AppendNotFound("driveFileId", "drive file")
			}
			return driveFile, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByIdResult {
			return &it.GetDriveFileByIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFile) *it.GetDriveFileByIdResult {
			return &it.GetDriveFileByIdResult{HasData: true, Data: d}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) DownloadDriveFile(ctx crud.Context, query it.GetDriveFileByIdQuery) (d *domain.DriveFile, rc io.ReadCloser, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "download drive file"); e != nil {
			err = e
		}
	}()
	driveFile, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return nil, nil, nil
	}

	ioReader, err := this.storageAdapter.DownloadBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.StorageKey)
	ft.PanicOnErr(err)

	return driveFile, ioReader, nil
}

// Get by parent

func (this *DriveFileServiceImpl) GetDriveFileByParent(ctx crud.Context, query it.GetDriveFileByParentQuery) (result *it.GetDriveFileByParentResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file by parent"); e != nil {
			err = e
		}
	}()
	query.SetDefaults()
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileByParentResult{ClientError: vErrs.ToClientError()}, nil
	}

	// Assert parent exists when listing children of a specific folder (skip for root: empty parent id)
	if query.FileParentId != "" {
		vErrsAssert := ft.NewValidationErrors()
		parentModel := &domain.DriveFile{ModelBase: model.ModelBase{Id: &query.FileParentId}}

		parent, err := this.assertDriveFileExists(ctx, parentModel, &vErrsAssert)
		ft.PanicOnErr(err)

		if parent == nil {
			return &it.GetDriveFileByParentResult{ClientError: vErrsAssert.ToClientError()}, nil
		}
	}

	return crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.GetDriveFileByParentQuery, it.GetDriveFileByParentResult]{
		Action: "get drive files by parent",
		Query:  query,
		SetQueryDefaults: func(q *it.GetDriveFileByParentQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.GetDriveFileByParentQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
			return this.driveFileRepo.SearchByParent(ctx, it.SearchByParentParam{
				ParentFileId: q.FileParentId,
				Predicate:    predicate,
				Order:        order,
				Page:         *q.Page,
				Size:         *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByParentResult {
			return &it.GetDriveFileByParentResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.GetDriveFileByParentResult {
			return &it.GetDriveFileByParentResult{Data: paged, HasData: true}
		},
	})
}

// Move to trash
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
	if cmd.ParentFileRef != nil || driveFile.ParentDriveFileRef != nil || *cmd.ParentFileRef != *driveFile.ParentDriveFileRef {
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

	driveFile, err = this.driveFileRepo.Overwrite(ctx, driveFile, *driveFile.Etag)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
	}

	ft.PanicOnErr(tx.Commit())
	return &it.RestoreDriveFileResult{HasData: true, Data: driveFile}, nil
}

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

	driveFile, err = this.driveFileRepo.Overwrite(ctx, driveFile, *driveFile.Etag)
	if err != nil {
		tx.Rollback()
		ft.PanicOnErr(err)
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

// Search

func (this *DriveFileServiceImpl) SearchDriveFile(ctx crud.Context, query it.SearchDriveFileQuery) (result *it.SearchDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "search drive file"); e != nil {
			err = e
		}
	}()
	result, err = crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.SearchDriveFileQuery, it.SearchDriveFileResult]{
		Action: "search drive files",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchDriveFileQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.SearchDriveFileQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
			return this.driveFileRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{Data: paged, HasData: paged.Items != nil}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) GetDriveFileAncestors(ctx crud.Context, query it.GetDriveFileAncestorsQuery) (result *it.GetDriveFileAncestorsResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file ancestors"); e != nil {
			err = e
		}
	}()
	vErrs := ft.NewValidationErrors()
	driveFile, err := this.assertDriveFileExists(ctx, query.ToDomainModel(), &vErrs)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return &it.GetDriveFileAncestorsResult{ClientError: vErrs.ToClientError()}, nil
	}

	if driveFile.ParentDriveFileRef == nil {
		return &it.GetDriveFileAncestorsResult{HasData: true, Data: []*domain.DriveFile{driveFile}}, nil
	}

	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, *driveFile.Id)
	ft.PanicOnErr(err)

	path := make([]*domain.DriveFile, len(ancestors))
	copy(path, ancestors)
	lo.Reverse(path)
	return &it.GetDriveFileAncestorsResult{HasData: true, Data: path}, nil
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

// assert methods
// ----------------------------------------------------------------------------------------//

func (this *DriveFileServiceImpl) sanitizeDriveFile(d *domain.DriveFile) {
	if d != nil && d.Name != "" {
		d.Name = defense.SanitizePlainText(d.Name, true)
	}
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
		}

		if !parent.IsFolder {
			vErrs.Append("parentFileRef", "parent drive must be a folder")
		}

		if parent.Status != enum.DriveFileStatusActive {
			vErrs.Append("parentFileRef", "parent drive must be active")
		}
	}

	if vErrs.Count() > 0 {
		return nil
	}

	d.Status = enum.DriveFileStatusActive

	return nil
}
