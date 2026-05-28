package app

import (
	"errors"
	"io"

	"github.com/samber/lo"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
	driveFileAncestorIt "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_ancestor"
)

type DriveFileServiceImpl struct {
	logger logging.LoggerService
	config config.ConfigService

	permissionSvc it.DriveFilePermissionService

	driveFileRepo         it.DriveFileRepository
	driveFileAncestorRepo driveFileAncestorIt.DriveFileAncestorRepository

	storageAdapter filestorage.FileStorage
}

func NewDriveFileService(
	logger logging.LoggerService,
	config config.ConfigService,

	permissionSvc it.DriveFilePermissionService,

	driveFileRepo it.DriveFileRepository,
	driveFileAncestorRepo driveFileAncestorIt.DriveFileAncestorRepository,

	storageAdapter filestorage.FileStorage,
) it.DriveFileService {
	return &DriveFileServiceImpl{
		logger:                logger,
		config:                config,
		driveFileRepo:         driveFileRepo,
		driveFileAncestorRepo: driveFileAncestorRepo,
		permissionSvc:         permissionSvc,
		storageAdapter:        storageAdapter,
	}
}

func (this *DriveFileServiceImpl) CreateDriveFile(
	ctx corectx.Context, cmd it.CreateDriveFileCommand) (*it.CreateDriveFileResult, error) {

	if cmd.FileHeader != nil {
		size := cmd.FileHeader.Size
		cmd.DriveFile.SetSize(&size)
	}

	mime := extractMIME(cmd.File)
	cmd.DriveFile.SetMime(&mime)

	storageKey := filestorage.BuildObjectKey(filestorage.BuildObjectKeyParams{
		Feature: "drive",
		Name:    *cmd.DriveFile.GetName(),
	})
	cmd.DriveFile.SetStorageKey(&storageKey)

	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	ctx.SetDbTranx(tx)
	defer tx.Rollback()

	createRes, err := crud.Create(ctx, crud.CreateParam[domain.DriveFile, *domain.DriveFile]{
		Action:         "Create drive file",
		BaseRepoGetter: this.driveFileRepo,
		Data:           cmd.DriveFile,
	})
	if err != nil {
		return nil, err
	}

	if createRes.ClientErrors.Count() > 0 {
		return createRes, nil
	}

	driveFile := createRes.Data

	// create ancestor Rel for new file
	err = this.createAncestorsRelByParent(ctx, driveFile.GetId(), driveFile.GetParentFileRef(), &createRes.ClientErrors)
	if err != nil {
		return nil, err
	}

	if createRes.ClientErrors.Count() > 0 {
		return createRes, nil
	}

	err = this.storageAdapter.Put(ctx, lo.FromPtr(driveFile.GetStorageKey()), cmd.File, *driveFile.GetSize(), nil)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	cmd.File.Close()

	return createRes, nil
}

func (this *DriveFileServiceImpl) UpdateDriveFileMetadata(ctx corectx.Context, cmd it.UpdateDriveFileMetadataCommand) (*it.UpdateDriveFileResult, error) {
	return crud.Update(ctx, crud.UpdateParam[domain.DriveFile, *domain.DriveFile]{
		Action:       "Update drive file metadata",
		DbRepoGetter: this.driveFileRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, inputModel, foundModel *domain.DriveFile, vErrs *ft.ClientErrors) error {
			if !lo.FromPtr(foundModel.GetIsFolder()) {
				err := domain.DriveFileNameValidate(lo.FromPtr(inputModel.GetName()))
				if err != nil {
					vErrs.Append(*err)
				}
			}
			return nil
		},
	})
}

func (this *DriveFileServiceImpl) UpdateBulkDriveFileMetadata(ctx corectx.Context, cmd it.UpdateBulkDriveFileMetadataCommand) (*it.UpdateBulkDriveFileMetadataResult, error) {
	data := make([]domain.DriveFile, 0, len(cmd.DriveFiles))
	for _, item := range cmd.DriveFiles {
		fieldData := item.GetFieldData()
		df := domain.NewDriveFileFrom(fieldData)
		data = append(data, *df)
	}

	return crud.UpdateBulk(ctx, crud.UpdateBulkParam[domain.DriveFile, *domain.DriveFile, domain.DriveFile]{
		Action:         "Update drive file metadata",
		BaseRepoGetter: this.driveFileRepo,
		Data:           data,
		ValidateExtra: func(ctx corectx.Context, inputModel, foundModel *domain.DriveFile, vErrs *ft.ClientErrors) error {
			if !lo.FromPtr(foundModel.GetIsFolder()) {
				err := domain.DriveFileNameValidate(lo.FromPtr(inputModel.GetName()))
				if err != nil {
					vErrs.Append(*err)
				}
			}
			return nil
		},
	})
}

func (this *DriveFileServiceImpl) UpdateDriveFileContent(ctx corectx.Context, cmd it.UpdateDriveFileContentCommand) (*it.UpdateDriveFileResult, error) {
	updateRes := &it.UpdateDriveFileResult{}
	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SetDbTranx(tx)

	defer tx.Rollback()

	// Get Exists file to get object key
	getOneRes, err := crud.GetOne[domain.DriveFile](ctx, crud.GetOneParam{
		Action:       "Get drive file",
		DbRepoGetter: this.driveFileRepo,
		Query: dynamicmodel.GetOneQuery{
			Id:     lo.FromPtr(cmd.GetId()),
			Fields: []string{domain.DriveFileFieldStorageKey},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(getOneRes.ClientErrors) > 0 {
		updateRes.ClientErrors = getOneRes.ClientErrors
		return updateRes, nil
	}

	if !getOneRes.HasData {
		updateRes.ClientErrors.Append(*ft.NewNotFoundError("id"))
		return updateRes, nil
	}

	foundDriveFile := getOneRes.Data

	if cmd.FileHeader != nil {
		size := cmd.FileHeader.Size
		cmd.DriveFile.SetSize(&size)
	}

	mime := extractMIME(cmd.File)
	cmd.DriveFile.SetMime(&mime)

	// Update metadata
	updateRes, err = this.UpdateDriveFileMetadata(ctx, it.UpdateDriveFileMetadataCommand{
		DriveFile: cmd.DriveFile,
	})
	if err != nil {
		return nil, err
	}

	if len(updateRes.ClientErrors) > 0 {
		return updateRes, nil
	}

	if cmd.File != nil {
		// Recalculate size of parent
		updateParentRes, err := this.recalculateSizeOfParent(ctx, *foundDriveFile.GetParentFileRef(), cmd.FileHeader.Size, true)
		if err != nil {
			return nil, err
		}

		if len(updateParentRes.ClientErrors) > 0 {
			updateRes.ClientErrors = updateParentRes.ClientErrors
			return updateRes, nil
		}

		// Put new file by stored key (overwrite)
		err = this.storageAdapter.Put(ctx, lo.FromPtr(foundDriveFile.GetStorageKey()), cmd.File, cmd.FileHeader.Size, nil)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return updateRes, err
}

func (this *DriveFileServiceImpl) DeleteDriveFile(ctx corectx.Context, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	delRes := &it.DeleteDriveFileResult{}
	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	ctx.SetDbTranx(tx)
	defer tx.Rollback()

	getOneRes, err := crud.GetOne[domain.DriveFile](ctx, crud.GetOneParam{
		Action:       "Get drive file",
		DbRepoGetter: this.driveFileRepo,
		Query: dynamicmodel.GetOneQuery{
			Id:     lo.FromPtr(&cmd.DriveFileId),
			Fields: []string{domain.DriveFileFieldStorageKey, domain.DriveFileFieldIsFolder},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(getOneRes.ClientErrors) > 0 {
		delRes.ClientErrors = getOneRes.ClientErrors
		return delRes, nil
	}

	if !getOneRes.HasData {
		delRes.ClientErrors.Append(*ft.NewNotFoundError("id"))
		return delRes, nil
	}

	foundDriveFile := getOneRes.Data

	if lo.FromPtrOr(foundDriveFile.GetIsFolder(), true) {
	} else {
		delRes, err = this.deleteExistedDriveFile(ctx, foundDriveFile)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return delRes, nil
}

func (this *DriveFileServiceImpl) DeleteTrashedDriveFile(ctx corectx.Context) error {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) DownloadDriveFile(ctx corectx.Context, query it.GetDriveFileByIdQuery) (*domain.DriveFile, io.ReadCloser, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) GetDriveFileAncestors(ctx corectx.Context, query it.GetDriveFileAncestorsQuery) (*it.GetDriveFileAncestorsResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) GetDriveFileById(ctx corectx.Context, query it.GetDriveFileByIdQuery) (*it.GetDriveFileByIdResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) GetDriveFileByParent(ctx corectx.Context, query it.GetDriveFileByParentQuery) (*it.GetDriveFileByParentResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) MoveDriveFile(ctx corectx.Context, cmd it.MoveDriveFileCommand) (*it.MoveDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) MoveDriveFileToTrash(ctx corectx.Context, cmd it.MoveDriveFileToTrashCommand) (*it.MoveDriveFileToTrashResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) RestoreDriveFile(ctx corectx.Context, cmd it.RestoreDriveFileCommand) (*it.RestoreDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) SearchDriveFile(ctx corectx.Context, query it.SearchDriveFileQuery) (*it.SearchDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) SearchDriveFilesShared(ctx corectx.Context, query it.SearchDriveFilesSharedQuery) (*it.SearchDriveFilesSharedResult, error) {
	panic("unimplemented")
}

// Create ancestors for file by file's parent
func (this *DriveFileServiceImpl) createAncestorsRelByParent(
	ctx corectx.Context,
	fileId *model.Id,
	parentId *model.Id,
	vErrs *ft.ClientErrors,
) error {
	if parentId == nil || fileId == nil {
		return nil
	}

	g := dmodel.NewSearchGraph()
	g.And(*dmodel.NewSearchNode().NewCondition(domain.DriveFileAncestorFieldFileRef, dmodel.Equals, *parentId))

	// Get parent's ancestor
	res, err := this.driveFileAncestorRepo.Search(ctx, dynamicmodel.RepoSearchParam{
		Page:  0,
		Size:  0,
		Graph: g,
	})
	if err != nil {
		return nil
	}

	if res.ClientErrors.Count() > 0 {
		vErrs.Append(res.ClientErrors...)
		return nil
	}

	if !res.HasData || len(res.Data.Items) == 0 {
		return nil
	}

	pAncestors := res.Data.Items

	// create ancestor for file by replacing domain.FileRef by fileId
	fileAncestors := make([]domain.DriveFileAncestor, 0, len(pAncestors)+1)
	for _, pA := range pAncestors {
		fA := domain.NewDriveFileAncestorFrom(pA.GetFieldData())
		fA.SetFileRef(fileId)

		depth := lo.FromPtr(fA.GetDepth()) + 1
		fA.SetDepth(&depth)

		fileAncestors = append(fileAncestors, *fA)
	}

	// Create rel (fileId, parentId)
	fParentRel := domain.NewDriveFileAncestor()
	fParentRel.SetFileRef(fileId)
	fParentRel.SetAncestorRef(parentId)

	depth := int64(1)
	fParentRel.SetDepth(&depth)

	insertRes, err := this.driveFileAncestorRepo.InsertBulk(ctx, fileAncestors)
	if err != nil {
		return nil
	}

	if insertRes.ClientErrors.Count() > 0 {
		vErrs.Append(insertRes.ClientErrors...)
		return nil
	}

	return nil
}

func (this *DriveFileServiceImpl) recalculateSizeOfParent(ctx corectx.Context,
	parentId model.Id, sizeDelta int64, inc bool) (*it.UpdateDriveFileResult, error) {
	driveFiles, err := this.driveFileRepo.GetDriveFileParents(ctx, parentId)
	if err != nil {
		return nil, err
	}

	for _, f := range driveFiles {
		var newSize int64
		if inc {
			newSize = *f.GetSize() + sizeDelta
		} else {
			newSize = *f.GetSize() - sizeDelta
		}
		f.SetSize(&newSize)
	}

	return crud.UpdateBulk(ctx, crud.UpdateBulkParam[domain.DriveFile, *domain.DriveFile, *domain.DriveFile]{
		Action:         "Update drive file metadata",
		BaseRepoGetter: this.driveFileRepo,
		Data:           driveFiles,
	})
}

// Delete existed drive file has IsFolder = true
func (this *DriveFileServiceImpl) deleteExistedDriveFile(ctx corectx.Context, driveFile domain.DriveFile) (*it.DeleteDriveFileResult, error) {
	if lo.FromPtrOr(driveFile.GetIsFolder(), true) {
		return nil, errors.New("driveFile must be a file")
	}

	err := this.storageAdapter.Remove(ctx, lo.FromPtr(driveFile.GetStorageKey()))
	if err != nil {
		return nil, err
	}

	return this.driveFileRepo.DeleteOne(ctx, driveFile)
}

// func (this *DriveFileServiceImpl) deleteDriveFileFolder(ctx corectx.Context, driveFile domain.DriveFile) (*it.DeleteDriveFileResult, error) {
// 	delRes := &it.DeleteDriveFileResult{}
//
// 	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
// 	ft.PanicOnErr(err)
// 	driveFile.BuildTree(children)
//
// 	folderChildren := []*domain.DriveFile{}
// 	fileChildren := []*domain.DriveFile{}
// 	for _, child := range children {
// 		if child.IsFolder {
// 			folderChildren = append(folderChildren, child)
// 		} else {
// 			fileChildren = append(fileChildren, child)
// 		}
// 	}
//
// 	storageKeys := lo.Map(fileChildren, func(driveFile *domain.DriveFile, index int) string {
// 		return driveFile.StorageKey
// 	})
//
// 	deletedKeys, _, err := this.storageAdapter.DeleteBulk(ctx, storageKeys)
// 	if err != nil {
// 		this.logger.Error("[DriveFileService] this.storageAdapter.DeleteBulk error", err)
// 		ft.PanicOnErr(err)
// 	}
//
// 	deletedKeySet := collections.NewSet(deletedKeys)
// 	deletedDriveFileIds := make([]string, 0, len(deletedKeys)+len(folderChildren))
// 	for _, child := range children {
// 		if deletedKeySet.Has(child.StorageKey) {
// 			deletedDriveFileIds = append(deletedDriveFileIds, *child.Id)
// 		}
// 	}
// 	for _, child := range folderChildren {
// 		deletedDriveFileIds = append(deletedDriveFileIds, *child.Id)
// 	}
//
// 	deletedDriveFileIdSet := collections.NewSet(deletedDriveFileIds)
// 	failedDriveFileIdSet := collections.NewSet([]model.Id{})
//
// 	// post-order: folder chỉ được xóa nếu toàn bộ children xóa thành công
// 	var postOrderDelete func(driveFile *domain.DriveFile) bool
// 	postOrderDelete = func(driveFile *domain.DriveFile) bool {
// 		shouldDelete := true
// 		for _, child := range driveFile.Children {
// 			shouldDelete = shouldDelete && postOrderDelete(child)
// 		}
// 		shouldDelete = shouldDelete && deletedDriveFileIdSet.Has(*driveFile.Id)
// 		if !shouldDelete {
// 			deletedDriveFileIdSet.Remove(*driveFile.Id)
// 			failedDriveFileIdSet.Add(*driveFile.Id)
// 		}
// 		return shouldDelete
// 	}
// 	postOrderDelete(driveFile)
//
// 	_, err = this.driveFileRepo.DeleteByIds(ctx, deletedDriveFileIdSet.GetValues())
// 	ft.PanicOnErr(err)
//
// 	// update status pending-delete cho các file/folder xóa thất bại
// 	allFiles := append(children, driveFile)
// 	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, failedDriveFileIdSet.Len())
// 	for _, f := range allFiles {
// 		if failedDriveFileIdSet.Has(*f.Id) {
// 			updateCmds = append(updateCmds, it.UpdateDriveFileMetadataCommand{
// 				Id:     *f.Id,
// 				Etag:   *f.Etag,
// 				Status: enum.DriveFileStatusPendingDelete,
// 			})
// 		}
// 	}
//
// 	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
// 		DriveFiles: updateCmds,
// 	})
// 	ft.PanicOnErr(err)
// 	return delRes, nil
// }
