package services

import (
	"io"
	"maps"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// Upload refuses everything it can before the first byte reaches storage: the declared metadata,
// the space, the quota and the content's signature. Only the exact size has to wait for the end
// of the stream.
func (this *FileDomainServiceImpl) Upload(
	ctx corectx.Context, cmd itFile.UploadFileCommand,
) (*itFile.UploadFileResult, error) {
	if cErrs := this.validateUpload(cmd); cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, nil
	}
	space, cErrs, err := this.resolveUploadSpace(ctx, cmd.OrgId, cmd.SpaceId)
	if err != nil || cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}
	if cErrs, err = this.assertQuota(ctx, cmd.SizeBytes); err != nil || cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}
	body, cErrs, err := checkContentMime(cmd)
	if err != nil || cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}

	fileId, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "Upload: new file id")
	}
	ext := models.ExtensionOf(cmd.Name)
	objectKey := models.BuildObjectKey(actingTenantId(ctx), util.ValueOrZeroOf(space.GetId()), *fileId, ext)
	if cErrs, err = this.putObject(ctx, objectKey, body, cmd); err != nil || cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}

	created, cErrs, err := this.recordUpload(ctx, cmd, space, *fileId, objectKey)
	if err != nil || cErrs.Count() > 0 {
		this.RemoveObjects(ctx, []string{objectKey})
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}
	created.FilePath = models.BuildFilePath(space.GetPath(), *fileId, ext)
	return &itFile.UploadFileResult{HasData: true, Data: *created}, nil
}

func (this *FileDomainServiceImpl) validateUpload(cmd itFile.UploadFileCommand) ft.ClientErrors {
	cErrs := ft.ClientErrors{}
	if cmd.Body == nil {
		cErrs.Append(*ft.NewValidationError(c.FieldFilePart, ft.ErrorKey("err_missing_required_field"),
			"the file part is required"))
	}
	if cmd.SizeBytes < 0 {
		cErrs.Append(*ft.NewValidationError(models.FileFieldSizeBytes, ft.ErrorKey("err_invalid_value"),
			"size_bytes must not be negative"))
	}
	if cmd.SizeBytes > this.maxFileSize {
		cErrs.Append(*ft.NewValidationError(models.FileFieldSizeBytes, ft.ErrorKey("err_file_too_large"),
			"file size {{actual_size}} exceeds the maximum of {{max_size}} bytes",
			map[string]any{"max_size": this.maxFileSize, "actual_size": cmd.SizeBytes}))
	}
	return cErrs
}

// resolveUploadSpace defaults to the org's space, then the tenant root. A named space must sit
// inside the org's space for an org file, and outside every org's for a tenant-wide one, so a
// caller cannot drop a file into another org's tree by naming its space.
func (this *FileDomainServiceImpl) resolveUploadSpace(
	ctx corectx.Context, orgId *model.Id, spaceId *model.Id,
) (*models.Space, ft.ClientErrors, error) {
	if spaceId != nil && *spaceId != "" {
		return this.resolveNamedSpace(ctx, orgId, *spaceId)
	}
	var space *models.Space
	var err error
	if orgId != nil {
		space, err = this.spaceSvc.FindOrgSpace(ctx, *orgId)
	} else {
		space, err = this.spaceSvc.FindTenantRootSpace(ctx)
	}
	if err != nil || space != nil {
		return space, nil, err
	}
	return nil, ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
		ft.ErrorKey("err_no_storage_space", c.ItStorageModuleName),
		"there is no storage space to upload into"),
	}, nil
}

func (this *FileDomainServiceImpl) resolveNamedSpace(
	ctx corectx.Context, orgId *model.Id, spaceId model.Id,
) (*models.Space, ft.ClientErrors, error) {
	space, err := this.spaceSvc.FindSpace(ctx, spaceId)
	if err != nil {
		return nil, nil, err
	}
	if space == nil {
		return nil, ft.ClientErrors{*ft.NewNotFoundError("space_id")}, nil
	}
	allowed := space.GetOwnerType() != c.OwnerTypeOrg
	if orgId != nil {
		orgSpace, err := this.spaceSvc.FindOrgSpace(ctx, *orgId)
		if err != nil {
			return nil, nil, err
		}
		allowed = orgSpace != nil && orgSpace.ContainsPath(space.GetPath())
	}
	if !allowed {
		return nil, ft.ClientErrors{*ft.NewValidationError("space_id",
			ft.ErrorKey("err_space_outside_scope", c.ItStorageModuleName),
			"the space does not belong to the organization the file is uploaded for")}, nil
	}
	return space, nil, nil
}

// assertQuota compares the tenant root's rolled-up size with the tenant's quota. A binary without
// tenants has neither, so it is not limited. A tenant without a quota row is refused: an
// unconfigured cap must not read as an unlimited one.
func (this *FileDomainServiceImpl) assertQuota(ctx corectx.Context, sizeBytes int64) (ft.ClientErrors, error) {
	tenantId := actingTenantId(ctx)
	if tenantId == "" {
		return nil, nil
	}
	quota, err := this.quotaSvc.GetQuota(ctx, itExt.GetQuotaQuery{
		Name: c.QuotaNameItStorage, OwnerType: c.OwnerTypeTenant, OwnerId: tenantId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Upload: read the storage quota")
	}
	if !quota.HasData || quota.Data.ItStorageCapacityMibytes == nil {
		return ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_quota_not_configured", c.ItStorageModuleName),
			"no storage quota is configured for this tenant")}, nil
	}
	root, err := this.spaceSvc.FindTenantRootSpace(ctx)
	if err != nil {
		return nil, err
	}
	usedBytes := int64(0)
	if root != nil {
		usedBytes = root.GetSizeBytes()
	}
	capacityBytes := *quota.Data.ItStorageCapacityMibytes * c.MebiByte
	if usedBytes+sizeBytes > capacityBytes {
		return ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_quota_exceeded", c.ItStorageModuleName),
			"the upload of {{size_bytes}} bytes exceeds the storage quota: {{used_bytes}} of {{capacity_bytes}} used",
			map[string]any{"size_bytes": sizeBytes, "used_bytes": usedBytes, "capacity_bytes": capacityBytes})}, nil
	}
	return nil, nil
}

// checkContentMime does not trust the declared type alone: the content's signature must agree
// with its family. An empty file has no signature to check.
func checkContentMime(cmd itFile.UploadFileCommand) (io.Reader, ft.ClientErrors, error) {
	if cmd.SizeBytes == 0 {
		return cmd.Body, nil, nil
	}
	detected, reader, err := sniffMime(cmd.Body)
	if err != nil {
		return nil, nil, err
	}
	if !models.MimeFamilyMatches(cmd.MimeType, detected) {
		return nil, ft.ClientErrors{*ft.NewValidationError(models.FileFieldMimeType,
			ft.ErrorKey("err_mime_type_mismatch", c.ItStorageModuleName),
			"the content is {{detected}}, which does not match the declared {{declared}}",
			map[string]any{"declared": cmd.MimeType, "detected": detected})}, nil
	}
	return reader, nil, nil
}

func (this *FileDomainServiceImpl) putObject(
	ctx corectx.Context, objectKey string, body io.Reader, cmd itFile.UploadFileCommand,
) (ft.ClientErrors, error) {
	reader := newExactSizeReader(body, cmd.SizeBytes)
	err := this.storage.Put(ctx, objectKey, reader, &filestorage.PutOptions{
		ContentLength: &cmd.SizeBytes,
		ContentType:   &cmd.MimeType,
	})
	if err == nil && reader.verifyDrained() {
		return nil, nil
	}
	if err != nil && !reader.mismatch {
		return nil, errors.Wrap(err, "Upload: store the object")
	}
	this.RemoveObjects(ctx, []string{objectKey})
	return ft.ClientErrors{*ft.NewValidationError(models.FileFieldSizeBytes,
		ft.ErrorKey("err_size_mismatch", c.ItStorageModuleName), errSizeMismatch.Error())}, nil
}

// recordUpload writes the row and adds its size to every enclosing space in one transaction.
func (this *FileDomainServiceImpl) recordUpload(
	ctx corectx.Context, cmd itFile.UploadFileCommand, space *models.Space, fileId model.Id, objectKey string,
) (*itFile.UploadFileResultData, ft.ClientErrors, error) {
	fields := dmodel.DynamicFields{
		models.FileFieldId:        fileId,
		models.FileFieldName:      cmd.Name,
		models.FileFieldMimeType:  cmd.MimeType,
		models.FileFieldSizeBytes: cmd.SizeBytes,
		models.FileFieldSpaceId:   util.ValueOrZeroOf(space.GetId()),
		models.FileFieldObjectKey: objectKey,
	}
	if cmd.OrgId != nil {
		fields[models.FileFieldOrgId] = *cmd.OrgId
	}
	var created *composable.CreateResult
	cErrs, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			created, err = this.CrudDomainService.Create(tranxCtx, fields, composable.CreateOptions{
				AfterValidationSuccess: pinFileId(fileId),
			})
			if err != nil {
				return nil, err
			}
			if created.ClientErrors.Count() > 0 {
				return &created.ClientErrors, nil
			}
			return nil, this.spaceSvc.AddSize(tranxCtx, util.ValueOrZeroOf(space.GetId()), cmd.SizeBytes)
		})
	if err != nil || (cErrs != nil && cErrs.Count() > 0) {
		return nil, util.ValueOrZeroOf(cErrs), err
	}
	return &itFile.UploadFileResultData{Created: created.Data}, nil, nil
}

// pinFileId restores the id the object key was built from. Schema validation regenerates every
// auto-generated field on create, the primary key included, but the object is already stored under
// this id and file_path must name it.
func pinFileId(fileId model.Id) composable.AfterValidationFn {
	return func(_ corectx.Context, entity *composable.DynamicEntity) (*composable.DynamicEntity, error) {
		fields := maps.Clone(entity.GetFieldData())
		fields[models.FileFieldId] = fileId
		return composable.NewDynamicEntityFrom(fields), nil
	}
}
