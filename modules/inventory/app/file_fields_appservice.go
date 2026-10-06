package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const paramOrgId = "org_id"

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
	stored, vErrs, err := filefield.StoreUploads(ctx, this.storage, cmd)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: *vErrs}, nil
	}

	created, err := this.CrudApplicationService.Create(ctx, cmd)
	if err != nil || created == nil || created.ClientErrors.Count() > 0 {
		filefield.RemoveObjects(ctx, this.storage, stored)
	}
	return created, err
}

func (this *fileBackedApplicationService) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	stored, vErrs, err := filefield.StoreUploads(ctx, this.storage, cmd)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *vErrs}, nil
	}

	updated, err := this.CrudApplicationService.Update(ctx, cmd)
	if err != nil || updated == nil || updated.ClientErrors.Count() > 0 {
		filefield.RemoveObjects(ctx, this.storage, stored)
	}
	return updated, err
}

func (this *fileBackedApplicationService) ClearFiles(
	ctx corectx.Context, recordId model.Id, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, params); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}

	fields, vErrs := filefield.ReadClearFileParams(params, this.fields)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	params[basemodel.FieldId] = string(recordId)
	current, err := this.CrudApplicationService.GetById(ctx, this.recordQuery(params))
	if err != nil {
		return nil, err
	}
	if current == nil || !current.HasData {
		return &dyn.OpResult[dyn.MutateResultData]{}, nil
	}

	update := this.recordQuery(params)
	update[basemodel.FieldEtag] = params[basemodel.FieldEtag]
	staleKeys := make([]string, 0, len(fields))
	for _, field := range fields {
		if key := current.Data.Item.GetString(field.KeyField); key != nil && *key != "" {
			staleKeys = append(staleKeys, *key)
		}
		update[field.KeyField] = nil
	}

	result, err := this.CrudApplicationService.Update(ctx, update)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return result, nil
	}

	filefield.RemoveObjects(ctx, this.storage, staleKeys)
	return result, nil
}

func (this *fileBackedApplicationService) recordQuery(params dmodel.DynamicFields) dmodel.DynamicFields {
	query := dmodel.DynamicFields{basemodel.FieldId: params[basemodel.FieldId]}
	if orgId, ok := params[paramOrgId]; ok {
		query[paramOrgId] = orgId
	}
	return query
}
