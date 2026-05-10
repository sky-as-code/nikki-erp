package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

func NewModelMetadataApplicationServiceImpl(modelMetadataSvc it.ModelMetadataDomainService) it.ModelMetadataAppService {
	return &ModelMetadataApplicationServiceImpl{modelMetadataSvc: modelMetadataSvc}
}

type ModelMetadataApplicationServiceImpl struct {
	modelMetadataSvc it.ModelMetadataDomainService
}

func (this *ModelMetadataApplicationServiceImpl) CreateModelMetadata(ctx corectx.Context, cmd it.CreateModelMetadataCommand) (*it.CreateModelMetadataResult, error) {
	return this.modelMetadataSvc.CreateModelMetadata(ctx, cmd)
}

func (this *ModelMetadataApplicationServiceImpl) DeleteModelMetadata(ctx corectx.Context, cmd it.DeleteModelMetadataCommand) (*it.DeleteModelMetadataResult, error) {
	return this.modelMetadataSvc.DeleteModelMetadata(ctx, cmd)
}

func (this *ModelMetadataApplicationServiceImpl) ModelMetadataExists(ctx corectx.Context, query it.ModelMetadataExistsQuery) (*it.ModelMetadataExistsResult, error) {
	return this.modelMetadataSvc.ModelMetadataExists(ctx, query)
}

func (this *ModelMetadataApplicationServiceImpl) GetModelMetadata(ctx corectx.Context, query it.GetModelMetadataQuery) (*it.GetModelMetadataResult, error) {
	return this.modelMetadataSvc.GetModelMetadata(ctx, query)
}

func (this *ModelMetadataApplicationServiceImpl) SearchModelMetadata(ctx corectx.Context, query it.SearchModelMetadataQuery) (*it.SearchModelMetadataResult, error) {
	return this.modelMetadataSvc.SearchModelMetadata(ctx, query)
}

func (this *ModelMetadataApplicationServiceImpl) UpdateModelMetadata(ctx corectx.Context, cmd it.UpdateModelMetadataCommand) (*it.UpdateModelMetadataResult, error) {
	return this.modelMetadataSvc.UpdateModelMetadata(ctx, cmd)
}
