package filefield

import (
	"mime/multipart"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type UploadFileCommand struct {
	RecordId  model.Id
	FieldName string
	File      *multipart.FileHeader
	Params    dmodel.DynamicFields
}

type DeleteFileCommand struct {
	RecordId  model.Id
	FieldName string
	Params    dmodel.DynamicFields
}

type FileFieldResult = dyn.OpResult[dyn.MutateResultData]

type FileFieldsService interface {
	UploadFile(ctx corectx.Context, cmd UploadFileCommand) (*FileFieldResult, error)
	ReplaceFile(ctx corectx.Context, cmd UploadFileCommand) (*FileFieldResult, error)
	DeleteFile(ctx corectx.Context, cmd DeleteFileCommand) (*FileFieldResult, error)
}

// ApplicationService wraps a resource's composable application service: reads presign every file
// field, writes refuse the managed key columns, and the file methods are the keys' only writer.
type ApplicationService struct {
	composable.CrudApplicationService

	storage filestorage.FileStorageAdapter
	fields  []FileField
}

// Storage is the adapter the service stores with, for a resource that presigns a key of its own
// outside the record maps (a projected view, say).
func (this *ApplicationService) Storage() filestorage.FileStorageAdapter {
	return this.storage
}

// Fields lists the resource's file fields.
func (this *ApplicationService) Fields() []FileField {
	return this.fields
}

func NewApplicationService(
	base composable.CrudApplicationService, storage filestorage.FileStorageAdapter, fields []FileField,
) ApplicationService {
	return ApplicationService{CrudApplicationService: base, storage: storage, fields: fields}
}

func (this *ApplicationService) GetById(
	ctx corectx.Context, query composable.GetByIdQuery,
) (*composable.GetOneResult, error) {
	return this.readOne(ctx, query, this.CrudApplicationService.GetById)
}

func (this *ApplicationService) GetOne(
	ctx corectx.Context, query composable.GetOneQuery,
) (*composable.GetOneResult, error) {
	return this.readOne(ctx, query, this.CrudApplicationService.GetOne)
}

func (this *ApplicationService) readOne(
	ctx corectx.Context, query dmodel.DynamicFields,
	read func(corectx.Context, dmodel.DynamicFields) (*composable.GetOneResult, error),
) (*composable.GetOneResult, error) {
	added := appendHiddenKeys(query, this.fields, nil)
	result, err := read(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if err := PresignFileFields(ctx, this.storage, result.Data.Item, this.fields); err != nil {
		return nil, err
	}
	meta := &result.Data.Meta
	meta.DesiredFields = stripHiddenKeys([]dmodel.DynamicFields{result.Data.Item}, meta.DesiredFields, added)
	return result, nil
}

// Search presigns every row, also when the projection is the default columns, which name the URL
// field's fallback rather than the storage key.
func (this *ApplicationService) Search(
	ctx corectx.Context, query composable.SearchQuery,
) (*composable.SearchResult, error) {
	added := appendHiddenKeys(query, this.fields, this.defaultProjection(query))
	result, err := this.CrudApplicationService.Search(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if err := PresignPage(ctx, this.storage, result.Data.Items, this.fields); err != nil {
		return nil, err
	}
	result.Data.DesiredFields = stripHiddenKeys(result.Data.Items, result.Data.DesiredFields, added)
	return result, nil
}

// defaultProjection is what a search without `fields` projects: the schema's default search
// fields for the default view. Any other named view resolves its own columns, left untouched.
func (this *ApplicationService) defaultProjection(query dmodel.DynamicFields) []string {
	if name, ok := query[basemodel.FieldSearchName].(string); ok && name != "" && name != dyn.DefaultSearchName {
		return nil
	}
	return this.Schema().DefaultSearchFields()
}

func (this *ApplicationService) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	if vErrs := ManagedFieldViolations(fieldNames(cmd), this.fields); vErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Create(ctx, cmd)
}

func (this *ApplicationService) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	if vErrs := ManagedFieldViolations(fieldNames(cmd), this.fields); vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Update(ctx, cmd)
}

func (this *ApplicationService) CreateBulk(
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
	if vErrs := ManagedFieldViolations(names, this.fields); vErrs.Count() > 0 {
		return &composable.BulkCreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.CreateBulk(ctx, cmd)
}

func (this *ApplicationService) Import(
	ctx corectx.Context, cmd composable.ImportCommand,
) (*composable.BulkCreateResult, error) {
	names := make([]string, 0, len(cmd.Mapping.Columns))
	for _, column := range cmd.Mapping.Columns {
		names = append(names, column.Target)
	}
	if vErrs := ManagedFieldViolations(names, this.fields); vErrs.Count() > 0 {
		return &composable.BulkCreateResult{ClientErrors: *vErrs}, nil
	}
	return this.CrudApplicationService.Import(ctx, cmd)
}

func (this *ApplicationService) UploadFile(
	ctx corectx.Context, cmd UploadFileCommand,
) (*FileFieldResult, error) {
	return this.storeFile(ctx, cmd, false)
}

func (this *ApplicationService) ReplaceFile(
	ctx corectx.Context, cmd UploadFileCommand,
) (*FileFieldResult, error) {
	return this.storeFile(ctx, cmd, true)
}

func (this *ApplicationService) storeFile(
	ctx corectx.Context, cmd UploadFileCommand, replace bool,
) (*FileFieldResult, error) {
	params := recordParams(cmd.RecordId, cmd.Params)
	if cErrs, err := assertRecordAction(this.CrudApplicationService, ctx, composable.PermissionUpdate, params); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}

	field, cErrs := this.resolveField(cmd.FieldName)
	if cErrs != nil {
		return mutateFailure(cErrs, nil)
	}
	if cmd.File == nil {
		return mutateFailure(newFileViolation("file", ft.ErrorKey("file.err_file_required"),
			"a 'file' part is required"), nil)
	}

	current, err := this.currentRecord(ctx, params)
	if err != nil || current == nil {
		return &FileFieldResult{}, err
	}

	oldKey := storedKey(current, field)
	if oldKey != "" && !replace {
		return mutateFailure(newFileViolation(field.Name, ft.ErrorKey("file.err_file_exists"),
			"'"+field.Name+"' already has a file; replace it with PUT"), nil)
	}

	vErrs := ft.NewClientErrors()
	newKey, mime, err := StoreFile(ctx, this.storage, field, cmd.File, vErrs)
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
		RemoveObjects(ctx, this.storage, []string{newKey})
		return result, err
	}

	if oldKey != "" {
		RemoveObjects(ctx, this.storage, []string{oldKey})
	}
	return result, nil
}

func (this *ApplicationService) DeleteFile(
	ctx corectx.Context, cmd DeleteFileCommand,
) (*FileFieldResult, error) {
	params := recordParams(cmd.RecordId, cmd.Params)
	if cErrs, err := assertRecordAction(this.CrudApplicationService, ctx, composable.PermissionUpdate, params); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}

	field, cErrs := this.resolveField(cmd.FieldName)
	if cErrs != nil {
		return mutateFailure(cErrs, nil)
	}

	current, err := this.currentRecord(ctx, params)
	if err != nil || current == nil {
		return &FileFieldResult{}, err
	}

	oldKey := storedKey(current, field)
	if oldKey == "" {
		return &FileFieldResult{
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

	RemoveObjects(ctx, this.storage, []string{oldKey})
	return result, nil
}

func (this *ApplicationService) resolveField(name string) (FileField, *ft.ClientErrors) {
	field, known := FindFileField(this.fields, name)
	if !known {
		return field, newFileViolation("field", ft.ErrorKey("file.err_field_unknown"),
			"'"+name+"' is not a file field of this resource")
	}
	return field, nil
}

func (this *ApplicationService) currentRecord(
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

func storedKey(record dmodel.DynamicFields, field FileField) string {
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

func assertRecordAction(
	appSvc composable.CrudApplicationService, ctx corectx.Context, permission string, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orgId, cErrs := appSvc.AssertAction(ctx, permission, params)
	if cErrs != nil {
		return cErrs, nil
	}
	if orgId == nil {
		return nil, nil
	}
	return appSvc.AssertRecordInOrg(ctx, params, *orgId)
}

func mutateFailure(cErrs *ft.ClientErrors, err error) (*FileFieldResult, error) {
	if err != nil {
		return nil, err
	}
	return &FileFieldResult{ClientErrors: *cErrs}, nil
}
