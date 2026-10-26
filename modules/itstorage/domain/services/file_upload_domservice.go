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
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// Upload refuses everything it can before the first byte reaches storage: the declared metadata,
// the space, the quota and the content's signature. Only the exact size has to wait for the end
// of the stream. The quota is held as a reservation from then on, and every failure gives it back.
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
	reservation, cErrs, err := this.reserveQuota(ctx, cmd.OrgId, cmd.SizeBytes)
	if err != nil || cErrs.Count() > 0 {
		return &itFile.UploadFileResult{ClientErrors: cErrs}, err
	}
	result, err := this.storeUpload(ctx, cmd, space, *reservation)
	if err != nil || result.ClientErrors.Count() > 0 {
		this.consumptionSvc.Release(ctx, *reservation)
	}
	return result, err
}

// storeUpload streams the object, tagged pending, then records it; the caller releases the
// reservation when this does not succeed.
func (this *FileDomainServiceImpl) storeUpload(
	ctx corectx.Context, cmd itFile.UploadFileCommand, space *models.Space, reservation itConsumption.Reservation,
) (*itFile.UploadFileResult, error) {
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
	fields := dmodel.DynamicFields{
		models.FileFieldName:      cmd.Name,
		models.FileFieldMimeType:  cmd.MimeType,
		models.FileFieldSizeBytes: cmd.SizeBytes,
		models.FileFieldObjectKey: objectKey,
	}
	created, cErrs, err := this.recordFile(ctx, fields, cmd.OrgId, space, *fileId, reservation)
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

// reserveQuota reserves the size on the tenant, under its quota, and on the file's org and its
// uploader without a cap. A binary without tenants has no quota. A tenant without a quota row is
// refused: an unconfigured cap must not read as an unlimited one.
func (this *FileDomainServiceImpl) reserveQuota(
	ctx corectx.Context, orgId *model.Id, sizeBytes int64,
) (*itConsumption.Reservation, ft.ClientErrors, error) {
	capBytes, cErrs, err := this.tenantCapacity(ctx)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	owners, err := this.consumptionOwners(ctx, orgId, uploaderOf(ctx))
	if err != nil {
		return nil, nil, err
	}
	reservation := itConsumption.Reservation{Owners: owners, Bytes: sizeBytes, CapBytes: capBytes}
	reserved, err := this.consumptionSvc.Reserve(ctx, reservation)
	if err != nil {
		return nil, nil, err
	}
	if !reserved {
		return nil, ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_quota_exceeded", c.ItStorageModuleName),
			"storing {{size_bytes}} bytes would exceed the storage quota of {{capacity_bytes}} bytes",
			map[string]any{"size_bytes": sizeBytes, "capacity_bytes": util.ValueOrZeroOf(capBytes)})}, nil
	}
	return &reservation, nil, nil
}

// tenantCapacity is nil, without client errors, in a binary without tenants.
func (this *FileDomainServiceImpl) tenantCapacity(ctx corectx.Context) (*int64, ft.ClientErrors, error) {
	tenantId := actingTenantId(ctx)
	if tenantId == "" {
		return nil, nil, nil
	}
	quota, err := this.quotaSvc.GetQuota(ctx, itExt.GetQuotaQuery{
		Name: c.QuotaNameItStorage, OwnerType: c.OwnerTypeTenant, OwnerId: tenantId,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "read the storage quota")
	}
	if !quota.HasData || quota.Data.ItStorageCapacityMibytes == nil {
		return nil, ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
			ft.ErrorKey("err_quota_not_configured", c.ItStorageModuleName),
			"no storage quota is configured for this tenant")}, nil
	}
	capacityBytes := *quota.Data.ItStorageCapacityMibytes * c.MebiByte
	return &capacityBytes, nil, nil
}

// consumptionOwners lists the rows a file counts towards, the tenant first since it is the one
// capped. A row created now starts from what the owner's space already holds.
func (this *FileDomainServiceImpl) consumptionOwners(
	ctx corectx.Context, orgId *model.Id, uploader *model.Id,
) ([]models.ConsumptionOwner, error) {
	owners := []models.ConsumptionOwner{}
	if tenantId := actingTenantId(ctx); tenantId != "" {
		root, err := this.spaceSvc.FindTenantRootSpace(ctx)
		if err != nil {
			return nil, err
		}
		owners = append(owners, models.ConsumptionOwner{
			Type: c.OwnerTypeTenant, Id: model.Id(tenantId), InitialUsedBytes: sizeOf(root),
		})
	}
	if orgId != nil {
		orgSpace, err := this.spaceSvc.FindOrgSpace(ctx, *orgId)
		if err != nil {
			return nil, err
		}
		owners = append(owners, models.ConsumptionOwner{
			Type: c.OwnerTypeOrg, Id: *orgId, InitialUsedBytes: sizeOf(orgSpace),
		})
	}
	if uploader != nil {
		owners = append(owners, models.ConsumptionOwner{Type: c.OwnerTypeUser, Id: *uploader})
	}
	return owners, nil
}

// fileOwners lists the rows an existing file counts towards. They exist already, so no seed.
func fileOwners(ctx corectx.Context, file *models.File) []models.ConsumptionOwner {
	owners := []models.ConsumptionOwner{}
	if tenantId := actingTenantId(ctx); tenantId != "" {
		owners = append(owners, models.ConsumptionOwner{Type: c.OwnerTypeTenant, Id: model.Id(tenantId)})
	}
	if orgId := file.GetOrgId(); orgId != nil {
		owners = append(owners, models.ConsumptionOwner{Type: c.OwnerTypeOrg, Id: *orgId})
	}
	if uploader := file.GetUploadedBy(); uploader != nil {
		owners = append(owners, models.ConsumptionOwner{Type: c.OwnerTypeUser, Id: *uploader})
	}
	return owners
}

// uploaderOf is the acting user; a service principal stores files on nobody's account.
func uploaderOf(ctx corectx.Context) *model.Id {
	perms := ctx.GetPermissions()
	if perms.Principal.Kind == corectx.PrincipalKindService || perms.UserId == "" {
		return nil
	}
	userId := perms.UserId
	return &userId
}

func sizeOf(space *models.Space) int64 {
	if space == nil {
		return 0
	}
	return space.GetSizeBytes()
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
		Tags:          stateTags(storage.FileStatusPending),
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

// recordFile writes the pending row, adds its size to every enclosing space and settles the
// reservation, all in one transaction.
func (this *FileDomainServiceImpl) recordFile(
	ctx corectx.Context, fields dmodel.DynamicFields, orgId *model.Id, space *models.Space, fileId model.Id,
	reservation itConsumption.Reservation,
) (*itFile.UploadFileResultData, ft.ClientErrors, error) {
	fields[models.FileFieldId] = fileId
	fields[models.FileFieldSpaceId] = util.ValueOrZeroOf(space.GetId())
	fields[models.FileFieldStatus] = storage.FileStatusPending
	if orgId != nil {
		fields[models.FileFieldOrgId] = *orgId
	}
	if uploader := uploaderOf(ctx); uploader != nil {
		fields[models.FileFieldUploadedBy] = *uploader
	}
	var created *composable.CreateResult
	cErrs, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			created, err = this.CrudDomainService.Create(tranxCtx, fields, composable.CreateOptions{
				AfterValidationSuccess: pinFileId(fileId),
			})
			if err != nil || created.ClientErrors.Count() > 0 {
				return clientErrorsOf(created), err
			}
			if err = this.spaceSvc.AddSize(tranxCtx, util.ValueOrZeroOf(space.GetId()), reservation.Bytes); err != nil {
				return nil, err
			}
			return nil, this.consumptionSvc.Commit(tranxCtx, reservation)
		})
	if err != nil || (cErrs != nil && cErrs.Count() > 0) {
		return nil, util.ValueOrZeroOf(cErrs), err
	}
	return &itFile.UploadFileResultData{Created: created.Data}, nil, nil
}

func clientErrorsOf(result *composable.CreateResult) *ft.ClientErrors {
	if result == nil {
		return nil
	}
	return &result.ClientErrors
}

// stateTags is the object tag a storage lifecycle rule matches to expire pending objects.
func stateTags(status string) map[string]string {
	return map[string]string{c.ObjectTagState: status}
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
