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
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileServiceImpl struct {
	logger         logging.LoggerService
	driveFileRepo  it.DriveFileRepository
	storageAdapter file_storage.FileStorageAdapter
}

func NewDriveFileService(
	logger logging.LoggerService,
	driveFileRepo it.DriveFileRepository,
	storageAdapter file_storage.FileStorageAdapter,
) it.DriveFileService {
	return &DriveFileServiceImpl{
		logger:         logger,
		driveFileRepo:  driveFileRepo,
		storageAdapter: storageAdapter,
	}
}

// Create
func (this *DriveFileServiceImpl) CreateDriveFile(ctx crud.Context, cmd it.CreateDriveFileCommand) (
	*it.CreateDriveFileResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.DriveFile, *it.CreateDriveFileCommand, it.CreateDriveFileResult]{
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
	return result, err
}

func (this *DriveFileServiceImpl) setCreateDriveFileDefaults(d *domain.DriveFile) {
	d.SetDefaults()
	d.Path = d.BuildObjectKeyStorage()
	d.Status = enum.DriveFileStatusActive
	d.Storage = enum.DriveFileStorageS3
}

func (this *DriveFileServiceImpl) assertCreateDriveFileRules(ctx crud.Context, d *domain.DriveFile, vErrs *ft.ValidationErrors) error {
	var ownerRef model.Id
	if d.OwnerRef != nil {
		ownerRef = *d.OwnerRef
	}
	exists, err := this.driveFileRepo.ExistsByOwnerParentNameFolder(ctx, ownerRef, d.ParentDriveFileRef, d.Name, d.IsFolder)
	if err != nil {
		return err
	}
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

	err = this.storageAdapter.UploadBucket(ctx.InnerContext(), file_storage.BucketDrive, d.Path, d.File)
	if err != nil {
		this.logger.Error("create drive file: storage upload failed", err)
		return err
	}
	return nil
}

// Update metadata

func (this *DriveFileServiceImpl) UpdateDriveFileMetadata(ctx crud.Context, cmd it.UpdateDriveFileMetadataCommand) (*it.UpdateDriveFileMetadataResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.DriveFile, it.UpdateDriveFileMetadataCommand, it.UpdateDriveFileMetadataResult]{
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

func (this *DriveFileServiceImpl) UpdateBulkDriveFileMetadata(ctx crud.Context, cmd it.UpdateBulkDriveFileMetadataCommand) (*it.UpdateBulkDriveFileMetadataResult, error) {
	result, err := crud.UpdateBulk(ctx, crud.UpdateBulkParam[*domain.DriveFile, it.UpdateBulkDriveFileMetadataCommand, it.UpdateBulkDriveFileMetadataResult]{
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

func (this *DriveFileServiceImpl) UpdateDriveFileContent(ctx crud.Context, cmd it.UpdateDriveFileContentCommand) (*it.UpdateDriveFileContentResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.DriveFile, it.UpdateDriveFileContentCommand, it.UpdateDriveFileContentResult]{
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

func (this *DriveFileServiceImpl) DeleteDriveFile(ctx crud.Context, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	var err error
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "check authorization"); e != nil {
			err = e
		}
	}()

	toFailureResultFunc := func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileResult {
		return &it.DeleteDriveFileResult{ClientError: vErrs.ToClientError()}
	}

	vErrs := util.ToPtr(fault.NewValidationErrors())
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), vErrs)
	if err != nil {
		return nil, err
	}

	if vErrs.Count() > 0 {
		return toFailureResultFunc(vErrs), nil
	}

	if !driveFile.IsFolder {
		return this.deleteExistsFile(ctx, driveFile, cmd)
	}

	err = this.deleteDriveFolder(ctx, driveFile)
	ft.PanicOnErr(err)

	return nil, nil
}

func (this *DriveFileServiceImpl) deleteExistsFile(ctx crud.Context, driveFile *domain.DriveFile, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	err := this.storageAdapter.DeleteBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.Path)
	if err != nil {
		this.logger.Error("delete drive file: storage delete failed", err)
		return nil, err
	}

	result, err := crud.DeleteHard(ctx,
		crud.DeleteHardParam[*domain.DriveFile, it.DeleteDriveFileCommand, it.DeleteDriveFileResult]{
			Action:  "delete drive file",
			Command: cmd,
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
	return result, err
}

func (this *DriveFileServiceImpl) deleteDriveFolder(ctx crud.Context, driveFile *domain.DriveFile) error {
	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
	if err != nil {
		return err
	}
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

	objectKeys := lo.Map(fileChildren, func(driveFile *domain.DriveFile, index int) string {
		return driveFile.Path
	})

	deletedKeys, _, err := this.storageAdapter.DeleteBulk(ctx, objectKeys)
	this.logger.Error("[DriveFileService] this.storageAdapter.DeleteBulk error", err)

	deletedKeySet := collections.NewSet(deletedKeys)
	deletedDriveFileIds := make([]string, 0, len(deletedKeys)+len(folderChildren))
	for _, child := range children {
		if deletedKeySet.Has(child.Path) {
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

func (this *DriveFileServiceImpl) GetDriveFileById(ctx crud.Context, query it.GetDriveFileByIdQuery) (*it.GetDriveFileByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.DriveFile, it.GetDriveFileByIdQuery, it.GetDriveFileByIdResult]{
		Action: "get drive file by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetDriveFileByIdQuery, vErrs *ft.ValidationErrors) (*domain.DriveFile, error) {
			driveFile, err := this.driveFileRepo.FindById(ctx, q.DriveFileId)
			if err != nil {
				return nil, err
			}
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

func (this *DriveFileServiceImpl) DownloadDriveFile(ctx crud.Context, query it.GetDriveFileByIdQuery) (*domain.DriveFile, io.ReadCloser, error) {
	driveFile, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	if err != nil {
		return nil, nil, err
	}

	if driveFile == nil {
		return nil, nil, nil
	}

	ioReader, err := this.storageAdapter.DownloadBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.Path)

	return driveFile, ioReader, nil
}

// Get by parent

func (this *DriveFileServiceImpl) GetDriveFileByParent(ctx crud.Context, query it.GetDriveFileByParentQuery) (*it.GetDriveFileByParentResult, error) {
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
		if err != nil {
			return nil, err
		}

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
func (this *DriveFileServiceImpl) MoveDriveFileToTrash(ctx crud.Context, cmd it.MoveDriveFileToTrashCommand) (*it.MoveDriveFileToTrashResult, error) {
	tx, err := this.driveFileRepo.BeginTransaction(ctx)

	ctx.SetDbTranx(tx)

	vErrs := ft.NewValidationErrors()
	driveFile, err := this.assertDriveFileExists(ctx, cmd.ToDomainModel(), &vErrs)
	if err != nil {
		return nil, err
	}

	if driveFile == nil {
		return &it.MoveDriveFileToTrashResult{ClientError: vErrs.ToClientError()}, nil
	}

	now := time.Now()
	driveFile.DeletedAt = &now
	driveFile.Status = enum.DriveFileStatusInTrash
	updated, err := this.driveFileRepo.Update(ctx, driveFile, *driveFile.Etag)
	if err != nil {
		return nil, err
	}

	if driveFile.IsFolder {
		children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, len(children))
		for _, f := range children {
			updateCmds = append(updateCmds, it.UpdateDriveFileMetadataCommand{
				Id:     *f.Id,
				Etag:   *f.Etag,
				Status: enum.DriveFileStatusParentInTrash,
			})
		}

		_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
			DriveFiles: updateCmds,
		})
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return &it.MoveDriveFileToTrashResult{HasData: true, Data: updated}, nil
}

// Search

func (this *DriveFileServiceImpl) SearchDriveFile(ctx crud.Context, query it.SearchDriveFileQuery) (*it.SearchDriveFileResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.SearchDriveFileQuery, it.SearchDriveFileResult]{
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

func (this *DriveFileServiceImpl) DeleteTrashedDriveFile(ctx crud.Context) error {
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
	if err != nil {
		return nil, err
	}

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
	d.Path = d.BuildObjectKeyStorage()

	if err := this.storageAdapter.DeleteBucket(ctx.InnerContext(), file_storage.BucketDrive, fromDb.Path); err != nil {
		this.logger.Error("update drive file content: storage delete failed", err)
		return err
	}

	if err := this.storageAdapter.UploadBucket(ctx.InnerContext(), file_storage.BucketDrive, d.Path, d.File); err != nil {
		this.logger.Error("update drive file content: storage upload failed", err)
		return err
	}

	return nil
}
