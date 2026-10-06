package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type fileBackedApplicationService struct {
	composable.CrudApplicationService

	storage filestorage.FileStorageAdapter
	fields  []filefield.FileField
}

func newFileBackedApplicationService(
	base composable.CrudApplicationService, storage filestorage.FileStorageAdapter, fields []filefield.FileField,
) fileBackedApplicationService {
	return fileBackedApplicationService{CrudApplicationService: base, storage: storage, fields: fields}
}

func (this *fileBackedApplicationService) GetById(
	ctx corectx.Context, query composable.GetByIdQuery,
) (*composable.GetOneResult, error) {
	result, err := this.CrudApplicationService.GetById(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if err := filefield.PresignFileFields(ctx, this.storage, result.Data.Item, this.fields); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *fileBackedApplicationService) GetOne(
	ctx corectx.Context, query composable.GetOneQuery,
) (*composable.GetOneResult, error) {
	result, err := this.CrudApplicationService.GetOne(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if err := filefield.PresignFileFields(ctx, this.storage, result.Data.Item, this.fields); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *fileBackedApplicationService) Search(
	ctx corectx.Context, query composable.SearchQuery,
) (*composable.SearchResult, error) {
	result, err := this.CrudApplicationService.Search(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if err := filefield.PresignPage(ctx, this.storage, result.Data.Items, this.fields); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *fileBackedApplicationService) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	if vErrs := filefield.ManagedFieldViolations(fieldNames(cmd), this.fields); vErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Create(ctx, cmd)
}

func (this *fileBackedApplicationService) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	if vErrs := filefield.ManagedFieldViolations(fieldNames(cmd), this.fields); vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Update(ctx, cmd)
}

func (this *fileBackedApplicationService) CreateBulk(
	ctx corectx.Context, cmd composable.BulkCreateCommand,
) (*composable.BulkCreateResult, error) {
	names := []string{}
	if items, ok := cmd["items"].([]any); ok {
		for _, item := range items {
			if row, ok := item.(map[string]any); ok {
				names = append(names, fieldNames(row)...)
			}
		}
	}
	if vErrs := filefield.ManagedFieldViolations(names, this.fields); vErrs.Count() > 0 {
		return &composable.BulkCreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.CreateBulk(ctx, cmd)
}

func (this *fileBackedApplicationService) Import(
	ctx corectx.Context, cmd composable.ImportCommand,
) (*composable.BulkCreateResult, error) {
	names := make([]string, 0, len(cmd.Mapping.Columns))
	for _, column := range cmd.Mapping.Columns {
		names = append(names, column.Target)
	}
	if vErrs := filefield.ManagedFieldViolations(names, this.fields); vErrs.Count() > 0 {
		return &composable.BulkCreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Import(ctx, cmd)
}

func (this *fileBackedApplicationService) UploadFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	return this.storeFile(ctx, cmd, false)
}

func (this *fileBackedApplicationService) ReplaceFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	return this.storeFile(ctx, cmd, true)
}

func (this *fileBackedApplicationService) storeFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand, replace bool,
) (*itProduct.FileFieldResult, error) {
	params := recordParams(cmd.RecordId, cmd.Params)
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, params); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}

	field, cErrs := this.resolveField(cmd.FieldName)
	if cErrs != nil {
		return mutateFailure(cErrs, nil)
	}
	if cmd.File == nil {
		return mutateFailure(newFileViolation(filefield.FormPartFile, "file.err_file_required",
			"a 'file' part is required"), nil)
	}

	current, err := this.currentRecord(ctx, params)
	if err != nil || current == nil {
		return &itProduct.FileFieldResult{}, err
	}

	oldKey := storedKey(current, field)
	if oldKey != "" && !replace {
		return mutateFailure(newFileViolation(field.Name, "file.err_file_exists",
			"'"+field.Name+"' already has a file; replace it with PUT"), nil)
	}

	vErrs := ft.NewClientErrors()
	newKey, mime, err := filefield.StoreFile(ctx, this.storage, field, cmd.File, vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return mutateFailure(vErrs, nil)
	}

	update := fileUpdateParams(params, current)
	update[field.KeyField] = newKey
	if field.MimeField != "" {
		update[field.MimeField] = mime
	}

	result, err := this.CrudApplicationService.Update(ctx, update)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		filefield.RemoveObjects(ctx, this.storage, []string{newKey})
		return result, err
	}

	if oldKey != "" {
		filefield.RemoveObjects(ctx, this.storage, []string{oldKey})
	}
	return result, nil
}

func (this *fileBackedApplicationService) DeleteFile(
	ctx corectx.Context, cmd itProduct.DeleteFileCommand,
) (*itProduct.FileFieldResult, error) {
	params := recordParams(cmd.RecordId, cmd.Params)
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, params); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}

	field, cErrs := this.resolveField(cmd.FieldName)
	if cErrs != nil {
		return mutateFailure(cErrs, nil)
	}

	current, err := this.currentRecord(ctx, params)
	if err != nil || current == nil {
		return &itProduct.FileFieldResult{}, err
	}

	oldKey := storedKey(current, field)
	if oldKey == "" {
		return &itProduct.FileFieldResult{
			Data:    dyn.MutateResultData{Etag: etagOf(current)},
			HasData: true,
		}, nil
	}

	update := fileUpdateParams(params, current)
	update[field.KeyField] = nil
	if field.MimeField != "" {
		update[field.MimeField] = nil
	}

	result, err := this.CrudApplicationService.Update(ctx, update)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}

	filefield.RemoveObjects(ctx, this.storage, []string{oldKey})
	return result, nil
}

func (this *fileBackedApplicationService) resolveField(name string) (filefield.FileField, *ft.ClientErrors) {
	field, known := filefield.FindFileField(this.fields, name)
	if !known {
		return field, newFileViolation(filefield.PathParamField, "file.err_field_unknown",
			"'"+name+"' is not a file field of this resource")
	}
	return field, nil
}

func (this *fileBackedApplicationService) currentRecord(
	ctx corectx.Context, params dmodel.DynamicFields,
) (dmodel.DynamicFields, error) {
	query := composable.GetByIdQuery{basemodel.FieldId: params[basemodel.FieldId]}
	if orgId, ok := params[basemodel.FieldOrgId]; ok {
		query[basemodel.FieldOrgId] = orgId
	}

	found, err := this.CrudApplicationService.GetById(ctx, query)
	if err != nil || found == nil || !found.HasData {
		return nil, err
	}
	return found.Data.Item, nil
}

func fieldNames(params map[string]any) []string {
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	return names
}

func recordParams(recordId model.Id, params dmodel.DynamicFields) dmodel.DynamicFields {
	bound := dmodel.DynamicFields{}
	for name, value := range params {
		bound[name] = value
	}
	bound[basemodel.FieldId] = string(recordId)
	return bound
}

func fileUpdateParams(params dmodel.DynamicFields, current dmodel.DynamicFields) dmodel.DynamicFields {
	update := dmodel.DynamicFields{basemodel.FieldId: params[basemodel.FieldId]}
	if orgId, ok := params[basemodel.FieldOrgId]; ok {
		update[basemodel.FieldOrgId] = orgId
	}

	etag := params.GetString(basemodel.FieldEtag)
	if etag == nil || *etag == "" {
		update[basemodel.FieldEtag] = string(etagOf(current))
	} else {
		update[basemodel.FieldEtag] = *etag
	}
	return update
}

func storedKey(record dmodel.DynamicFields, field filefield.FileField) string {
	if key := record.GetString(field.KeyField); key != nil {
		return *key
	}
	return ""
}

func etagOf(record dmodel.DynamicFields) model.Etag {
	if etag := record.GetString(basemodel.FieldEtag); etag != nil {
		return model.Etag(*etag)
	}
	return ""
}

func newFileViolation(field, code, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewValidationError(field, code, message))
	return vErrs
}
