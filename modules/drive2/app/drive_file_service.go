package app

import (
	"errors"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/collections"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/objectkey"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
	itDriveFileAncestor "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_ancestor"
)

type DriveFileServiceImpl struct {
	logger logging.LoggerService
	config config.ConfigService

	permissionSvc            it.PermissionDomainService
	driveFileService         it.DriveFileDomainService
	driveFileAncestorService itDriveFileAncestor.DriveFileAncestorDomainService

	driveFileRepo it.DriveFileRepository

	storageAdapter filestorage.FileStorageAdapter
}

func NewDriveFileService(
	logger logging.LoggerService,
	config config.ConfigService,

	permissionSvc it.PermissionDomainService,
	driveFileService it.DriveFileDomainService,
	driveFileAncestorService itDriveFileAncestor.DriveFileAncestorDomainService,

	driveFileRepo it.DriveFileRepository,

	storageAdapter filestorage.FileStorageAdapter,
) it.DriveFileAppService {
	return &DriveFileServiceImpl{
		logger:                   logger,
		config:                   config,
		driveFileService:         driveFileService,
		driveFileAncestorService: driveFileAncestorService,
		driveFileRepo:            driveFileRepo,
		permissionSvc:            permissionSvc,
		storageAdapter:           storageAdapter,
	}
}

func (this *DriveFileServiceImpl) CreateDriveFile(
	ctx corectx.Context, cmd it.CreateDriveFileCommand) (*it.CreateDriveFileResult, error) {
	createRes := &it.CreateDriveFileResult{}

	if !lo.FromPtrOr(cmd.DriveFile.GetIsFolder(), true) {
		if cmd.FileHeader != nil {
			size := cmd.FileHeader.Size
			cmd.DriveFile.SetSize(&size)
		}

		mime := extractMIME(cmd.File)
		cmd.DriveFile.SetMime(&mime)

		storageKey, err := objectkey.BuildFromFileHeader("drive", cmd.FileHeader)
		if err != nil {
			return nil, err
		}

		cmd.DriveFile.SetStorageKey(&storageKey)
	}

	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SetDbTranx(tx)
	defer tx.Rollback()

	driveFile, err, ok := extractOpData(&createRes.ClientErrors, func() (*dynamicmodel.OpResult[models.DriveFile], error) {
		return this.driveFileService.CreateDriveFile(ctx, it.CreateDriveFileCommand{
			DriveFile: cmd.DriveFile,
		})
	}, models.DriveFileFieldId)
	if !ok {
		return createRes, err
	}

	// create ancestor Rel for new file
	err = this.createAncestorsRelByParent(ctx, driveFile.GetId(), driveFile.GetParentFileRef(), &createRes.ClientErrors)
	if err != nil {
		return nil, err
	}

	if createRes.ClientErrors.Count() > 0 {
		return createRes, nil
	}

	if !lo.FromPtrOr(cmd.DriveFile.GetIsFolder(), true) {
		err = this.storageAdapter.Put(ctx, lo.FromPtr(driveFile.GetStorageKey()),
			cmd.File,
			filestorage.NewPutOptions(lo.FromPtr(driveFile.GetMime()), lo.FromPtr(driveFile.GetSize())))
		if err != nil {
			return nil, err
		}

		cmd.File.Close()
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	createRes.HasData = true
	createRes.Data = driveFile

	return createRes, nil
}

func (this *DriveFileServiceImpl) UpdateDriveFileMetadata(ctx corectx.Context, cmd it.UpdateDriveFileMetadataCommand) (*it.UpdateDriveFileResult, error) {
	return this.driveFileService.UpdateDriveFileMetadata(ctx, cmd)
}

func (this *DriveFileServiceImpl) UpdateBulkDriveFileMetadata(ctx corectx.Context, cmd it.UpdateBulkDriveFileMetadataCommand) (*it.UpdateBulkDriveFileMetadataResult, error) {
	return this.driveFileService.UpdateBulkDriveFileMetadata(ctx, cmd)
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
	foundDriveFile, err, ok := extractOpData(&updateRes.ClientErrors,
		func() (*dynamicmodel.OpResult[models.DriveFile], error) {
			return this.driveFileService.GetDriveFileById(ctx, it.GetDriveFileByIdQuery{
				Id: lo.FromPtr(cmd.GetId()),
			})
		})
	if !ok {
		return updateRes, err
	}

	if cmd.FileHeader != nil {
		size := cmd.FileHeader.Size
		cmd.DriveFile.SetSize(&size)
	}

	mime := extractMIME(cmd.File)
	cmd.DriveFile.SetMime(&mime)

	updateRes, err, ok = resolveOpResult(func() (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
		return this.driveFileService.UpdateDriveFileMetadata(ctx, cmd.DriveFile)
	})
	if !ok {
		return updateRes, err
	}

	if cmd.File != nil {
		sizeDelta := cmd.FileHeader.Size - lo.FromPtr(foundDriveFile.GetSize())
		updateParentRes, err := this.recalculateSizeOfParent(ctx, lo.FromPtr(foundDriveFile.GetId()), sizeDelta)
		if err != nil {
			return nil, err
		}

		if len(updateParentRes.ClientErrors) > 0 {
			updateRes.ClientErrors = updateParentRes.ClientErrors
			return updateRes, nil
		}

		// Put new file by stored key (overwrite)
		err = this.storageAdapter.Put(ctx,
			lo.FromPtr(foundDriveFile.GetStorageKey()),
			cmd.File,
			filestorage.NewPutOptions(lo.FromPtr(cmd.DriveFile.GetMime()), lo.FromPtr(cmd.DriveFile.GetSize())))
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

	foundDriveFile, err, ok := extractOpData(&delRes.ClientErrors, func() (*dynamicmodel.OpResult[models.DriveFile], error) {
		return this.driveFileService.GetDriveFileById(ctx, it.GetDriveFileByIdQuery{
			Id:     lo.FromPtr(&cmd.DriveFileId),
			Fields: []string{models.DriveFileFieldStorageKey, models.DriveFileFieldIsFolder},
		})
	})
	if !ok {
		return delRes, err
	}

	_, err, ok = extractOpData(&delRes.ClientErrors, func() (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
		return this.recalculateSizeOfParent(ctx, lo.FromPtr(foundDriveFile.GetId()), lo.FromPtrOr(foundDriveFile.GetSize(), 0))
	})
	if !ok {
		return delRes, err
	}

	if lo.FromPtrOr(foundDriveFile.GetIsFolder(), true) {
		delRes, err = this.deleteDriveFileFolder(ctx, foundDriveFile)
	} else {
		delRes, err = this.deleteDriveFileFile(ctx, foundDriveFile)
	}

	if err != nil || delRes.ClientErrors.Count() > 0 {
		return delRes, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return delRes, nil
}

func (this *DriveFileServiceImpl) DeleteTrashedDriveFile(ctx corectx.Context) error {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) DownloadDriveFile(ctx corectx.Context, query it.DownloadDriveFileQuery) (res *it.DownloadDriveFileResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "download drive file"); e != nil {
			err = e
		}
	}()
	res = &it.DownloadDriveFileResult{}
	driveFile, err, ok := extractOpData(&res.ClientErrors, func() (*dynamicmodel.OpResult[models.DriveFile], error) {
		return this.driveFileService.GetDriveFileById(ctx, it.GetDriveFileByIdQuery{
			Id: query.Id,
		})
	})
	if !ok {
		return
	}

	if lo.FromPtrOr(driveFile.GetIsFolder(), true) {
		res.ClientErrors = ft.ClientErrors{*ft.NewValidationError("id", fault.ErrorKey("err_validation"), "file must be not a folder")}
	}

	// Permission check skipped here temporarily (public stream route); restore resolvePermission + CanView when auth/token is in place.

	openRes, err := this.storageAdapter.Open(ctx.InnerContext(), lo.FromPtr(driveFile.GetStorageKey()), "")
	ft.PanicOnErr(err)

	res.HasData = true
	res.Data.ContentLength = openRes.ContentLength
	res.Data.ContentRange = openRes.ContentRange
	res.Data.MineType = lo.FromPtr(driveFile.GetMime())
	res.Data.File = openRes.Body

	return
}

func (this *DriveFileServiceImpl) GetDriveFileAncestors(ctx corectx.Context, query it.GetDriveFileAncestorsQuery) (*it.GetDriveFileAncestorsResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) GetDriveFileById(ctx corectx.Context, query it.GetDriveFileByIdQuery) (*it.GetDriveFileByIdResult, error) {
	return crud.GetOne[models.DriveFile](ctx, crud.GetOneParam{
		Action:       "Get drive file by id",
		DbRepoGetter: this.driveFileRepo,
		Query: dynamicmodel.GetOneQuery{
			Id:     query.Id,
			Fields: query.Fields,
		},
	})
}

func (this *DriveFileServiceImpl) GetDriveFileByParent(ctx corectx.Context, query it.GetDriveFileByParentQuery) (*it.GetDriveFileByParentResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) MoveDriveFile(ctx corectx.Context, cmd it.MoveDriveFileCommand) (*it.MoveDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileServiceImpl) MoveDriveFileToTrash(ctx corectx.Context, cmd it.MoveDriveFileToTrashCommand) (*it.MoveDriveFileToTrashResult, error) {
	res := &it.MoveDriveFileToTrashResult{
		ClientErrors: ft.ClientErrors{},
		HasData:      false,
	}
	driveFile, err, ok := extractOpData(&res.ClientErrors, func() (*dynamicmodel.OpResult[models.DriveFile], error) {
		return this.driveFileService.GetDriveFileById(ctx, it.GetDriveFileByIdQuery{
			Id: cmd.DriveFileId,
		})
	})
	if !ok {
		return res, err
	}

	tx, err := this.driveFileRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	ctx.SetDbTranx(tx)
	defer tx.Rollback()

	status := models.DriveFileStatusInTrash
	driveFile.SetStatus(&status)

	res, err = this.driveFileService.UpdateDriveFileMetadata(ctx, driveFile)

	if lo.FromPtrOr(driveFile.GetIsFolder(), true) {
		data, err, ok := extractOpData(&res.ClientErrors, func() (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
			return this.updateChildrenStatue(ctx, cmd.DriveFileId, models.DriveFileStatusInTrash)
		})
		if !ok {
			return res, err
		}

		res.Data.AffectedCount += data.AffectedCount
	}

	if err := tx.Commit(); err != nil {
		return res, err
	}

	return res, nil
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
	g.And(*dmodel.NewSearchNode().NewCondition(models.DriveFileAncestorFieldFileRef, dmodel.Equals, *parentId))

	// Get parent's ancestor
	res, err := this.driveFileAncestorService.SearchDriveFileAncestors(ctx, itDriveFileAncestor.SearchDriveFileAncestorsQuery{
		Page:  0,
		Size:  0,
		Graph: g,
	})
	if err != nil {
		return err
	}

	if res.ClientErrors.Count() > 0 {
		vErrs.Append(res.ClientErrors...)
		return nil
	}

	pAncestors := []models.DriveFileAncestor{}
	if res.HasData && len(res.Data.Items) > 0 {
		pAncestors = res.Data.Items
	}

	fileAncestors := make([]models.DriveFileAncestor, 0, len(pAncestors)+1)
	for _, pA := range pAncestors {
		fA := models.NewDriveFileAncestor()
		fA.SetFileRef(fileId)
		fA.SetAncestorRef(pA.GetAncestorRef())

		depth := lo.FromPtr(pA.GetDepth()) + 1
		fA.SetDepth(&depth)

		fileAncestors = append(fileAncestors, *fA)
	}

	fParentRel := models.NewDriveFileAncestor()
	fParentRel.SetFileRef(fileId)
	fParentRel.SetAncestorRef(parentId)

	depth := int64(1)
	fParentRel.SetDepth(&depth)
	fileAncestors = append(fileAncestors, *fParentRel)

	insertRes, err := this.driveFileAncestorService.CreateBulkDriveFileAncestors(ctx, itDriveFileAncestor.CreateBulkDriveFileAncestorsCommand{
		Items: fileAncestors,
	})
	if err != nil {
		return err
	}

	if insertRes.ClientErrors.Count() > 0 {
		vErrs.Append(insertRes.ClientErrors...)
		return nil
	}

	return nil
}

func (this *DriveFileServiceImpl) recalculateSizeOfParent(
	ctx corectx.Context, childId model.Id, sizeDelta int64,
) (*it.UpdateDriveFileResult, error) {
	if sizeDelta == 0 {
		return &it.UpdateDriveFileResult{
			HasData: true,
			Data: dynamicmodel.MutateResultData{
				AffectedCount: 0,
			},
		}, nil
	}

	g := dmodel.NewSearchGraph()
	g.And(*dmodel.NewSearchNode().NewCondition(models.DriveFileAncestorFieldFileRef, dmodel.Equals, childId))

	searchRes, err := this.driveFileAncestorService.SearchDriveFileAncestors(ctx, itDriveFileAncestor.SearchDriveFileAncestorsQuery{
		Page:   0,
		Size:   0,
		Graph:  g,
		Fields: []string{models.DriveFileAncestorEdgeAncestorFile},
	})
	if err != nil {
		return nil, err
	}
	if searchRes.ClientErrors.Count() > 0 {
		return &it.UpdateDriveFileResult{ClientErrors: searchRes.ClientErrors}, nil
	}
	if !searchRes.HasData || len(searchRes.Data.Items) == 0 {
		return &it.UpdateDriveFileResult{}, nil
	}

	updateCommands := make([]it.UpdateDriveFileMetadataCommand, 0, len(searchRes.Data.Items))
	for _, ancestor := range searchRes.Data.Items {
		f := ancestor.GetAncestorFile()
		if f == nil || lo.FromPtr(f.GetId()) == "" {
			continue
		}

		newSize := lo.FromPtr(f.GetSize()) + sizeDelta
		cmd := it.UpdateDriveFileMetadataCommand{}
		cmd.SetSize(&newSize)
		cmd.SetId(f.GetId())
		cmd.SetEtag(f.GetEtag())
		updateCommands = append(updateCommands, cmd)
	}

	if len(updateCommands) == 0 {
		return &it.UpdateDriveFileResult{}, nil
	}

	return this.driveFileService.UpdateBulkDriveFileMetadata(ctx, updateCommands)
}

// Delete existed drive file has IsFolder = true
func (this *DriveFileServiceImpl) deleteDriveFileFile(ctx corectx.Context, driveFile models.DriveFile) (*it.DeleteDriveFileResult, error) {
	if lo.FromPtrOr(driveFile.GetIsFolder(), true) {
		return nil, errors.New("driveFile must be a file")
	}

	err := this.storageAdapter.Remove(ctx, lo.FromPtr(driveFile.GetStorageKey()))
	if err != nil {
		return nil, err
	}

	return this.driveFileService.DeleteDriveFile(ctx, it.DeleteDriveFileCommand{
		DriveFileId: lo.FromPtr(driveFile.GetId()),
	})
}

func (this *DriveFileServiceImpl) deleteDriveFileFolder(ctx corectx.Context, driveFile models.DriveFile) (*it.DeleteDriveFileResult, error) {
	delRes := &it.DeleteDriveFileResult{}

	childrenData, err, ok := extractOpData(&delRes.ClientErrors, func() (*dynamicmodel.OpResult[it.GetDriveFileChildrenResultData], error) {
		return this.driveFileService.GetDriveFileChildren(ctx, it.GetDriveFileChildrenQuery{
			DriveFileId: lo.FromPtr(driveFile.GetId()),
			Page:        0,
			Size:        model.MODEL_RULE_PAGE_MAX_SIZE,
		})
	})
	if !ok {
		return delRes, err
	}

	children := childrenData.Items

	driveFile.BuildTree(children)

	folderChildren := []*models.DriveFile{&driveFile}
	fileChildren := []*models.DriveFile{}
	for _, child := range children {
		if lo.FromPtr(child.GetIsFolder()) {
			folderChildren = append(folderChildren, child)
		} else {
			fileChildren = append(fileChildren, child)
		}
	}

	storageKeys := lo.Map(fileChildren, func(driveFile *models.DriveFile, index int) string {
		return lo.FromPtr(driveFile.GetStorageKey())
	})

	deletedKeys, _, err := this.storageAdapter.RemoveBulk(ctx, storageKeys)
	if err != nil {
		this.logger.Error("[DriveFileService] this.storageAdapter.DeleteBulk error", err)
		ft.PanicOnErr(err)
	}

	deletedKeySet := collections.NewSet(deletedKeys)
	deletedDriveFileIds := make([]string, 0, len(deletedKeys)+len(folderChildren))
	for _, child := range children {
		if deletedKeySet.Has(lo.FromPtr(child.GetStorageKey())) {
			deletedDriveFileIds = append(deletedDriveFileIds, lo.FromPtr(child.GetId()))
		}
	}
	for _, child := range folderChildren {
		deletedDriveFileIds = append(deletedDriveFileIds, lo.FromPtr(child.GetId()))
	}

	deletedDriveFileIdSet := collections.NewSet(deletedDriveFileIds)
	failedDriveFileIdSet := collections.NewSet([]model.Id{})

	// post-order: folder chỉ được xóa nếu toàn bộ children xóa thành công
	var postOrderDelete func(driveFile *models.DriveFile) bool
	postOrderDelete = func(driveFile *models.DriveFile) bool {
		shouldDelete := true
		driveFileChildren := driveFile.GetChildren()
		for _, child := range driveFileChildren {
			shouldDelete = shouldDelete && postOrderDelete(child)
		}
		shouldDelete = shouldDelete && deletedDriveFileIdSet.Has(lo.FromPtr(driveFile.GetId()))
		if !shouldDelete {
			deletedDriveFileIdSet.Remove(lo.FromPtr(driveFile.GetId()))
			failedDriveFileIdSet.Add(lo.FromPtr(driveFile.GetId()))
		}
		return shouldDelete
	}
	postOrderDelete(&driveFile)

	delRes, err, ok = resolveOpResult(func() (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
		return this.driveFileService.DeleteDriveFiles(ctx, it.DeleteDriveFilesCommand{
			DriveFileIds: deletedDriveFileIdSet.GetValues(),
		})
	})
	if !ok {
		return delRes, err
	}

	// update status pending-delete cho các file/folder xóa thất bại
	allFiles := append(children, &driveFile)
	updateCmds := make([]models.DriveFile, 0, failedDriveFileIdSet.Len())
	pendingDeleteStatus := models.DriveFileStatusPendingDelete
	for _, f := range allFiles {
		if failedDriveFileIdSet.Has(lo.FromPtr(f.GetId())) {
			f.SetStatus(&pendingDeleteStatus)
			updateCmds = append(updateCmds, *f)
		}
	}

	_, err, ok = extractOpData(&delRes.ClientErrors, func() (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
		return this.UpdateBulkDriveFileMetadata(ctx, updateCmds)
	})
	if !ok {
		return delRes, err
	}

	return delRes, nil
}

func (this *DriveFileServiceImpl) updateChildrenStatue(ctx corectx.Context,
	id model.Id,
	status models.DriveFileStatus) (*dynamicmodel.OpResult[dynamicmodel.MutateResultData], error) {
	updateRes := &dynamicmodel.OpResult[dynamicmodel.MutateResultData]{}
	childrenData, err, ok := extractOpData(&updateRes.ClientErrors, func() (*dynamicmodel.OpResult[it.GetDriveFileChildrenResultData], error) {
		return this.driveFileService.GetDriveFileChildren(ctx, it.GetDriveFileChildrenQuery{
			DriveFileId: id,
			Page:        0,
			Size:        model.MODEL_RULE_PAGE_MAX_SIZE,
		})
	})
	if !ok {
		return updateRes, err
	}

	children := childrenData.Items
	cmd := it.UpdateBulkDriveFileMetadataCommand{}
	for _, child := range children {
		child.SetStatus(&status)
		cmd = append(cmd, *child)
	}

	return this.UpdateBulkDriveFileMetadata(ctx, cmd)
}
