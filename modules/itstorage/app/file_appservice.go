package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// metaFields is what the meta API answers: the record without its storage internals.
var metaFields = []string{
	models.FileFieldId, models.FileFieldName, models.FileFieldMimeType, models.FileFieldSizeBytes,
	models.FileFieldCreatedAt, models.FileFieldSpaceId, models.FileFieldOrgId, models.FileFieldEtag,
	models.FileFieldStatus,
}

// NewFileApplicationService is handed the composable default by the itstorage_file onion.
func NewFileApplicationService(base composable.CrudApplicationService) itFile.FileApplicationService {
	return &FileApplicationServiceImpl{
		CrudApplicationService: base,
		domSvc:                 base.DomainService().(itFile.FileDomainService),
	}
}

type FileApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itFile.FileDomainService
}

// Create is refused: a row without its object would be a dangling file. Files arrive through
// UploadFile only.
func (this *FileApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	return &composable.CreateResult{ClientErrors: ft.ClientErrors{*ft.NewAnonymousBusinessViolation(
		ft.ErrorKey("err_action_not_supported"),
		"This resource does not support the requested action",
		map[string]any{"action": string(composable.CrudActionCreate), "resource": models.FileSchemaName},
	)}}, nil
}

func (this *FileApplicationServiceImpl) UploadFile(
	ctx corectx.Context, cmd itFile.UploadFileCommand,
) (*itFile.UploadFileResult, error) {
	if cErrs := assertFileAction(ctx, this, composable.PermissionCreate, cmd.OrgId); cErrs != nil {
		return &itFile.UploadFileResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.Upload(ctx, cmd)
}

func (this *FileApplicationServiceImpl) GetFileMeta(
	ctx corectx.Context, query itFile.GetFileMetaQuery,
) (*itFile.GetFileMetaResult, error) {
	resolved, cErrs, err := this.resolveReadable(ctx, query.Path)
	if err != nil || cErrs != nil {
		return &itFile.GetFileMetaResult{ClientErrors: valueOf(cErrs)}, err
	}
	stored := resolved.File.GetFieldData()
	meta := dmodel.DynamicFields{models.FileFieldFilePath: resolved.FilePath}
	for _, field := range metaFields {
		meta[field] = stored[field]
	}
	return &itFile.GetFileMetaResult{HasData: true, Data: meta}, nil
}

func (this *FileApplicationServiceImpl) DownloadFile(
	ctx corectx.Context, query itFile.DownloadFileQuery,
) (*itFile.DownloadFileResult, error) {
	resolved, cErrs, err := this.resolveReadable(ctx, query.Path)
	if err != nil || cErrs != nil {
		return &itFile.DownloadFileResult{ClientErrors: valueOf(cErrs)}, err
	}
	stream, err := this.domSvc.OpenObject(ctx, resolved.File, query.RangeHeader)
	if err != nil {
		return nil, err
	}
	return &itFile.DownloadFileResult{HasData: true, Data: *stream}, nil
}

// DownloadFileById is how this system's own clients read a file: by the id a record's file field
// holds. A missing id answers like one the caller may not see, after the tenant-wide read check.
func (this *FileApplicationServiceImpl) DownloadFileById(
	ctx corectx.Context, query itFile.DownloadFileByIdQuery,
) (*itFile.DownloadFileResult, error) {
	file, err := this.domSvc.FindFile(ctx, query.Id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		if cErrs := assertFileAction(ctx, this, composable.PermissionRead, nil); cErrs != nil {
			return &itFile.DownloadFileResult{ClientErrors: *cErrs}, nil
		}
		return &itFile.DownloadFileResult{ClientErrors: ft.ClientErrors{*ft.NewNotFoundError("id")}}, nil
	}
	if cErrs := assertFileRead(ctx, this, file.GetOrgId()); cErrs != nil {
		return &itFile.DownloadFileResult{ClientErrors: *cErrs}, nil
	}
	stream, err := this.domSvc.OpenObject(ctx, file, query.RangeHeader)
	if err != nil {
		return nil, err
	}
	return &itFile.DownloadFileResult{HasData: true, Data: *stream}, nil
}

// GetFileMetaById authorizes like DownloadFileById: against the org the stored record carries.
func (this *FileApplicationServiceImpl) GetFileMetaById(
	ctx corectx.Context, query itFile.GetFileMetaByIdQuery,
) (*itFile.GetFileMetaResult, error) {
	file, err := this.domSvc.FindFile(ctx, query.Id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		if cErrs := assertFileAction(ctx, this, composable.PermissionRead, nil); cErrs != nil {
			return &itFile.GetFileMetaResult{ClientErrors: *cErrs}, nil
		}
		return &itFile.GetFileMetaResult{ClientErrors: ft.ClientErrors{*ft.NewNotFoundError("id")}}, nil
	}
	if cErrs := assertFileRead(ctx, this, file.GetOrgId()); cErrs != nil {
		return &itFile.GetFileMetaResult{ClientErrors: *cErrs}, nil
	}
	stored := file.GetFieldData()
	meta := dmodel.DynamicFields{}
	for _, field := range metaFields {
		meta[field] = stored[field]
	}
	return &itFile.GetFileMetaResult{HasData: true, Data: meta}, nil
}

// resolveReadable authorizes against the org the stored record carries, never one the caller
// names. A path naming no file still asserts the tenant-wide read first, so the answer cannot be
// used to probe for files without the read grant.
func (this *FileApplicationServiceImpl) resolveReadable(
	ctx corectx.Context, filePath string,
) (*itFile.ResolvedFile, *ft.ClientErrors, error) {
	resolved, err := this.domSvc.FindByPath(ctx, filePath)
	if err != nil {
		return nil, nil, err
	}
	if resolved == nil {
		if cErrs := assertFileAction(ctx, this, composable.PermissionRead, nil); cErrs != nil {
			return nil, cErrs, nil
		}
		return nil, &ft.ClientErrors{*ft.NewNotFoundError("path")}, nil
	}
	if cErrs := assertFileRead(ctx, this, resolved.File.GetOrgId()); cErrs != nil {
		return nil, cErrs, nil
	}
	return resolved, nil, nil
}

func valueOf(cErrs *ft.ClientErrors) ft.ClientErrors {
	if cErrs == nil {
		return nil
	}
	return *cErrs
}
