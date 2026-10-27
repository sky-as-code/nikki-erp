package services

import (
	"maps"
	"slices"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

// releaseTagDelays is how long a release made inside the caller's transaction waits, attempt by
// attempt, for that transaction to commit before tagging the objects pending. Nothing signals the
// commit, so the committed rows are re-read; a release still unseen after the last attempt rolled
// back, or ran in a transaction longer than the wait, and its object keeps the in_use tag (a leak,
// never a loss: docs/problems itstorage-003).
var releaseTagDelays = []time.Duration{time.Second, 4 * time.Second, 15 * time.Second, time.Minute, 5 * time.Minute}

var fileStateFields = []string{
	models.FileFieldId, models.FileFieldOrgId, models.FileFieldStatus, models.FileFieldInUseAt,
	models.FileFieldObjectKey, models.FileFieldSizeBytes, models.FileFieldName, models.FileFieldMimeType,
	models.FileFieldSpaceId, models.FileFieldUploadedBy,
}

func (this *FileDomainServiceImpl) AssertFilesUsable(
	ctx corectx.Context, query storage.AssertFilesUsableQuery,
) (ft.ClientErrors, error) {
	files, err := this.findFiles(ctx, slices.Collect(maps.Keys(query.Fields)))
	if err != nil {
		return nil, err
	}
	cErrs := ft.ClientErrors{}
	for _, fileId := range slices.Sorted(maps.Keys(query.Fields)) {
		if reason := unusableReason(files[fileId], query.OrgId); reason != "" {
			cErrs.Append(*ft.NewValidationError(query.Fields[fileId],
				ft.ErrorKey("err_file_not_usable", c.ItStorageModuleName), reason))
		}
	}
	return cErrs, nil
}

func (this *FileDomainServiceImpl) GetFileMeta(
	ctx corectx.Context, query storage.GetFileMetaQuery,
) (*storage.FileMeta, ft.ClientErrors, error) {
	files, err := this.findFiles(ctx, []model.Id{query.FileId})
	if err != nil {
		return nil, nil, err
	}
	file := files[query.FileId]
	if file == nil || !sameOrg(file.GetOrgId(), query.OrgId) {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewValidationError(query.Field,
			ft.ErrorKey("err_file_not_usable", c.ItStorageModuleName), "no stored file of this organization has this id"))
		return nil, cErrs, nil
	}
	return &storage.FileMeta{MimeType: file.GetMimeType(), SizeBytes: file.GetSizeBytes()}, nil, nil
}

func unusableReason(file *models.File, orgId *model.Id) string {
	switch {
	case file == nil:
		return "no stored file has this id"
	case !sameOrg(file.GetOrgId(), orgId):
		return "the file belongs to another organization than the record referencing it"
	case file.GetStatus() != storage.FileStatusPending:
		return "the file is already referenced by another record; duplicate it instead"
	case file.IsReleased():
		return "the file was removed from the record that used it; upload it again"
	}
	return ""
}

func sameOrg(fileOrgId *model.Id, wantOrgId *model.Id) bool {
	if fileOrgId == nil || wantOrgId == nil {
		return fileOrgId == nil && wantOrgId == nil
	}
	return *fileOrgId == *wantOrgId
}

// MarkFilesInUse tags the objects in_use before it returns, so a tag failure aborts the caller's
// save: a record must never commit pointing at an object the lifecycle rule would expire. If the
// caller rolls back afterwards, the object keeps an in_use tag on a pending row, which only leaks.
//
// Every id must move. One that does not was taken by a concurrent write, was released, or never
// existed; refusing it is what keeps a file from being shared between two records.
func (this *FileDomainServiceImpl) MarkFilesInUse(ctx corectx.Context, cmd storage.MarkFilesInUseCommand) error {
	fileIds := uniqueIds(cmd.FileIds)
	moved, err := this.repo.SetStatus(ctx, fileIds, storage.FileStatusPending, storage.FileStatusInUse)
	if err != nil {
		return err
	}
	if len(moved) != len(fileIds) {
		missing := slices.DeleteFunc(slices.Clone(fileIds), func(id model.Id) bool { _, ok := moved[id]; return ok })
		return errors.Errorf("MarkFilesInUse: files %v are no longer pending", missing)
	}
	for _, objectKey := range moved {
		if err := this.storage.SetTags(ctx, objectKey, stateTags(storage.FileStatusInUse)); err != nil {
			return errors.Wrap(err, "MarkFilesInUse: tag the object in_use")
		}
	}
	return nil
}

// ReleaseFiles moves the rows back to pending in the caller's transaction, so a rollback keeps
// them in use. The objects are tagged pending only once that is committed: without a transaction
// at once, inside one by a detached re-check of the committed rows.
func (this *FileDomainServiceImpl) ReleaseFiles(ctx corectx.Context, cmd storage.ReleaseFilesCommand) error {
	moved, err := this.repo.SetStatus(ctx, uniqueIds(cmd.FileIds), storage.FileStatusInUse, storage.FileStatusPending)
	if err != nil || len(moved) == 0 {
		return err
	}
	if ctx.GetDbTranx() == nil {
		this.tagPending(ctx, moved)
		return nil
	}
	go this.tagPendingOnceCommitted(detachedContext(ctx), slices.Collect(maps.Keys(moved)))
	return nil
}

func (this *FileDomainServiceImpl) tagPendingOnceCommitted(ctx corectx.Context, fileIds []model.Id) {
	defer func() {
		if recovered := recover(); recovered != nil {
			ctx.GetLogger().Errorf("itstorage: tagging released files panicked: %v", recovered)
		}
	}()
	waiting := fileIds
	for _, delay := range releaseTagDelays {
		time.Sleep(delay)
		committed, err := this.repo.ListObjectKeysWithStatus(ctx, waiting, storage.FileStatusPending)
		if err != nil {
			ctx.GetLogger().Warn("itstorage: cannot re-read released files", logging.Attr{"error": err.Error()})
			continue
		}
		this.tagPending(ctx, committed)
		waiting = slices.DeleteFunc(waiting, func(id model.Id) bool { _, ok := committed[id]; return ok })
		if len(waiting) == 0 {
			return
		}
	}
	ctx.GetLogger().Info("itstorage: released files never committed as pending; their objects stay tagged in_use",
		logging.Attr{"file_ids": waiting})
}

// tagPending is best-effort: an object left tagged in_use only leaks.
func (this *FileDomainServiceImpl) tagPending(ctx corectx.Context, objectKeys map[model.Id]string) {
	for fileId, objectKey := range objectKeys {
		if err := this.storage.SetTags(ctx, objectKey, stateTags(storage.FileStatusPending)); err != nil {
			ctx.GetLogger().Warn("itstorage: failed to tag a released object pending; it will not expire",
				logging.Attr{"file_id": fileId, "object_key": objectKey, "error": err.Error()})
		}
	}
}

// DuplicateFile copies the object server-side into the target org's space and records the copy as
// a new pending file, reserved against the quota like an upload. Inside the caller's transaction,
// the row and the reservation commit or roll back with the caller's record.
func (this *FileDomainServiceImpl) DuplicateFile(
	ctx corectx.Context, cmd storage.DuplicateFileCommand,
) (*model.Id, error) {
	source, err := this.findFile(ctx, cmd.FileId)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, composable.ClientErrorSignal{Errors: ft.ClientErrors{*ft.NewNotFoundError("file_id")}}
	}
	space, cErrs, err := this.resolveUploadSpace(ctx, cmd.OrgId, nil)
	if err != nil || cErrs.Count() > 0 {
		return nil, signalOrError(cErrs, err)
	}
	reservation, cErrs, err := this.reserveQuota(ctx, cmd.OrgId, source.GetSizeBytes())
	if err != nil || cErrs.Count() > 0 {
		return nil, signalOrError(cErrs, err)
	}
	fileId, cErrs, err := this.copyFile(ctx, source, cmd.OrgId, space, *reservation)
	if err != nil || cErrs.Count() > 0 {
		this.consumptionSvc.Release(ctx, *reservation)
		return nil, signalOrError(cErrs, err)
	}
	return fileId, nil
}

func (this *FileDomainServiceImpl) copyFile(
	ctx corectx.Context, source *models.File, orgId *model.Id, space *models.Space, reservation itConsumption.Reservation,
) (*model.Id, ft.ClientErrors, error) {
	fileId, err := model.NewId()
	if err != nil {
		return nil, nil, errors.Wrap(err, "DuplicateFile: new file id")
	}
	spaceId := *space.GetId()
	objectKey := models.BuildObjectKey(actingTenantId(ctx), spaceId, *fileId, models.ExtensionOf(source.GetObjectKey()))
	err = this.storage.Copy(ctx, source.GetObjectKey(), objectKey, &filestorage.CopyOptions{
		Tags: stateTags(storage.FileStatusPending),
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "DuplicateFile: copy the object")
	}
	fields := dmodel.DynamicFields{
		models.FileFieldName:      source.GetName(),
		models.FileFieldMimeType:  source.GetMimeType(),
		models.FileFieldSizeBytes: source.GetSizeBytes(),
		models.FileFieldObjectKey: objectKey,
	}
	if _, cErrs, err := this.recordFile(ctx, fields, orgId, space, *fileId, reservation); err != nil || cErrs.Count() > 0 {
		this.RemoveObjects(ctx, []string{objectKey})
		return nil, cErrs, err
	}
	return fileId, nil, nil
}

// findFiles reads the files by id in one statement; ids naming no file are absent from the map.
func (this *FileDomainServiceImpl) findFiles(ctx corectx.Context, fileIds []model.Id) (map[model.Id]*models.File, error) {
	files := map[model.Id]*models.File{}
	if len(fileIds) == 0 {
		return files, nil
	}
	keys := make([]any, 0, len(fileIds))
	for _, fileId := range fileIds {
		keys = append(keys, string(fileId))
	}
	rows, err := composable.SearchRepositoryRows(
		ctx, this.Repository(), models.FileSchemaName, models.FileFieldId, keys, fileStateFields)
	if err != nil {
		return nil, errors.Wrap(err, "find files")
	}
	for _, row := range rows {
		file := models.NewFileFrom(row)
		if fileId := file.GetId(); fileId != nil {
			files[*fileId] = file
		}
	}
	return files, nil
}

// signalOrError hands client errors out of a port method that returns only an error, so a caller
// inside composable.InRefusableTransaction answers 400 instead of 500.
func signalOrError(cErrs ft.ClientErrors, err error) error {
	if err != nil {
		return err
	}
	return composable.ClientErrorSignal{Errors: cErrs}
}

func uniqueIds(ids []model.Id) []model.Id {
	unique := slices.Clone(ids)
	slices.Sort(unique)
	return slices.Compact(slices.DeleteFunc(unique, func(id model.Id) bool { return id == "" }))
}
