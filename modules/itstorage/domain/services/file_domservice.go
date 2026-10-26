package services

import (
	"strconv"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

// FileDomainDeps is what the file domain service needs beyond its composable default.
type FileDomainDeps struct {
	SpaceSvc       itSpace.SpaceDomainService
	Storage        filestorage.FileStorageAdapter
	QuotaSvc       itExt.QuotaExtService
	ConsumptionSvc itConsumption.ConsumptionDomainService
	ConfigSvc      config.ConfigService
}

// NewFileDomainService is handed the composable default by the itstorage_file onion.
func NewFileDomainService(base composable.CrudDomainService, deps FileDomainDeps) itFile.FileDomainService {
	maxSize := c.DefaultUploadMaxFileSizeBytes
	if deps.ConfigSvc != nil {
		// The configuration service parses its fallback from a string.
		maxSize = deps.ConfigSvc.GetInt64(c.UploadMaxFileSizeBytes, strconv.FormatInt(c.DefaultUploadMaxFileSizeBytes, 10))
	}
	return &FileDomainServiceImpl{
		CrudDomainService: base,
		repo:              base.Repository().(itFile.FileRepository),
		spaceSvc:          deps.SpaceSvc,
		storage:           deps.Storage,
		quotaSvc:          deps.QuotaSvc,
		consumptionSvc:    deps.ConsumptionSvc,
		maxFileSize:       maxSize,
	}
}

// FileDomainServiceImpl keeps a file row, its stored object and the size roll-up of its spaces in
// step. A row and the size change commit together; the object is written before and removed
// after, because storage cannot join the database transaction.
type FileDomainServiceImpl struct {
	composable.CrudDomainService
	repo           itFile.FileRepository
	spaceSvc       itSpace.SpaceDomainService
	storage        filestorage.FileStorageAdapter
	quotaSvc       itExt.QuotaExtService
	consumptionSvc itConsumption.ConsumptionDomainService
	maxFileSize    int64
}

// FindByPath matches the file's current path exactly. Internal references hold the file id; a path
// serves external apps only, and tracking a rename of the space is their responsibility.
func (this *FileDomainServiceImpl) FindByPath(ctx corectx.Context, filePath string) (*itFile.ResolvedFile, error) {
	parsed, ok := models.ParseFilePath(filePath)
	if !ok {
		return nil, nil
	}
	file, err := this.findFile(ctx, parsed.FileId)
	if err != nil || file == nil {
		return nil, err
	}
	space, err := this.spaceSvc.FindSpace(ctx, file.GetSpaceId())
	if err != nil {
		return nil, err
	}
	if space == nil {
		return nil, errors.Errorf("file %s points at missing space %s", parsed.FileId, file.GetSpaceId())
	}
	currentPath := models.BuildFilePath(space.GetPath(), parsed.FileId, models.ExtensionOf(file.GetObjectKey()))
	if currentPath != filePath {
		return nil, nil
	}
	return &itFile.ResolvedFile{File: file, Space: space, FilePath: currentPath}, nil
}

func (this *FileDomainServiceImpl) FindFile(ctx corectx.Context, fileId model.Id) (*models.File, error) {
	return this.findFile(ctx, fileId)
}

func (this *FileDomainServiceImpl) findFile(ctx corectx.Context, fileId model.Id) (*models.File, error) {
	if fileId == "" {
		return nil, nil
	}
	found, err := this.Repository().FindByKeys(ctx, dmodel.DynamicFields{models.FileFieldId: fileId})
	if err != nil {
		return nil, errors.Wrap(err, "find file")
	}
	if found.ClientErrors.Count() > 0 || !found.HasData {
		return nil, nil
	}
	return models.NewFileFrom(found.Data), nil
}

// Delete takes the file's size off its spaces and its consumption in the same transaction as the
// row, and removes the object once that commits. Inside a transaction the caller opened, the
// removal precedes the caller's commit, so a later rollback leaves a row without its object.
//
// A file in use is refused: the record referencing it would keep a dangling id. The record drops
// the reference instead, which releases the file to expire.
func (this *FileDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	file, err := this.findFile(ctx, composable.ParamId(cmd))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return this.CrudDomainService.Delete(ctx, cmd, options...)
	}
	if file.GetStatus() == storage.FileStatusInUse {
		return &composable.MutateResult{ClientErrors: ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_file_in_use", c.ItStorageModuleName),
			"the file is referenced by a record; remove it from that record instead")}}, nil
	}

	var result *composable.MutateResult
	cErrs, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			result, err = this.CrudDomainService.Delete(tranxCtx, cmd, options...)
			if err != nil {
				return nil, err
			}
			if result.ClientErrors.Count() > 0 {
				return &result.ClientErrors, nil
			}
			if err = this.spaceSvc.AddSize(tranxCtx, file.GetSpaceId(), -file.GetSizeBytes()); err != nil {
				return nil, err
			}
			return nil, this.consumptionSvc.Subtract(tranxCtx, fileOwners(tranxCtx, file), file.GetSizeBytes())
		})
	if err != nil {
		return nil, errors.Wrap(err, "FileDomainService.Delete")
	}
	if cErrs != nil && cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *cErrs}, nil
	}
	this.RemoveObjects(ctx, []string{file.GetObjectKey()})
	return result, nil
}

func (this *FileDomainServiceImpl) OpenObject(
	ctx corectx.Context, file *models.File, rangeHeader string,
) (*itFile.DownloadFileResultData, error) {
	stream, err := this.storage.Open(ctx, file.GetObjectKey(), rangeHeader)
	if err != nil {
		return nil, errors.Wrap(err, "open stored object")
	}
	return &itFile.DownloadFileResultData{
		Name:          file.GetName(),
		MimeType:      file.GetMimeType(),
		Body:          stream.Body,
		ContentLength: stream.ContentLength,
		ContentRange:  stream.ContentRange,
	}, nil
}

func (this *FileDomainServiceImpl) DeleteSpaceTree(ctx corectx.Context, space *models.Space) ([]string, error) {
	spaceRepo := this.spaceSvc.Repository().(itSpace.SpaceRepository)
	spaceIds, err := spaceRepo.ListSubtreeIds(ctx, space.GetPath())
	if err != nil {
		return nil, err
	}
	objectKeys, err := this.repo.ListObjectKeysInSpaces(ctx, spaceIds)
	if err != nil {
		return nil, err
	}
	if err := this.subtractTreeConsumption(ctx, space, spaceIds); err != nil {
		return nil, err
	}
	if _, err := this.repo.DeleteInSpaces(ctx, spaceIds); err != nil {
		return nil, err
	}
	if parentId := space.GetParentId(); parentId != nil {
		if err := this.spaceSvc.AddSize(ctx, *parentId, -space.GetSizeBytes()); err != nil {
			return nil, err
		}
	}
	if _, err := spaceRepo.DeleteSpaces(ctx, spaceIds); err != nil {
		return nil, err
	}
	return objectKeys, nil
}

// subtractTreeConsumption takes the tree's files off the tenant and their uploaders, and drops the
// row of the org that owned the tree: the org is going away.
func (this *FileDomainServiceImpl) subtractTreeConsumption(
	ctx corectx.Context, space *models.Space, spaceIds []model.Id,
) error {
	byUploader, err := this.repo.SumSizesByUploader(ctx, spaceIds)
	if err != nil {
		return err
	}
	if err := this.consumptionSvc.SubtractEach(ctx, c.OwnerTypeUser, byUploader); err != nil {
		return err
	}
	if tenantId := actingTenantId(ctx); tenantId != "" {
		tenant := models.ConsumptionOwner{Type: c.OwnerTypeTenant, Id: model.Id(tenantId)}
		if err := this.consumptionSvc.Subtract(ctx, []models.ConsumptionOwner{tenant}, space.GetSizeBytes()); err != nil {
			return err
		}
	}
	if ownerId := space.GetOwnerId(); ownerId != nil && space.GetOwnerType() == c.OwnerTypeOrg {
		return this.consumptionSvc.DeleteOwner(ctx, models.ConsumptionOwner{Type: c.OwnerTypeOrg, Id: *ownerId})
	}
	return nil
}

// RemoveObjects never fails the caller: the rows are already gone, and an object left behind is
// an orphan to reconcile, not a reason to undo (docs/problems itstorage-003).
func (this *FileDomainServiceImpl) RemoveObjects(ctx corectx.Context, objectKeys []string) {
	for _, objectKey := range objectKeys {
		if err := this.storage.Remove(ctx, objectKey); err != nil {
			ctx.GetLogger().Warn("itstorage: failed to remove a stored object; it is now an orphan", logging.Attr{
				"object_key": objectKey,
				"error":      err.Error(),
			})
		}
	}
}
