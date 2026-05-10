package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

func NewModelMetadataDomainServiceImpl(repo it.ModelMetadataRepository) it.ModelMetadataDomainService {
	return &ModelMetadataDomainServiceImpl{repo: repo}
}

type ModelMetadataDomainServiceImpl struct{ repo it.ModelMetadataRepository }

func (this *ModelMetadataDomainServiceImpl) CreateModelMetadata(ctx corectx.Context, cmd it.CreateModelMetadataCommand) (*it.CreateModelMetadataResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.ModelMetadata, *models.ModelMetadata]{
		Action: "create model metadata", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *ModelMetadataDomainServiceImpl) DeleteModelMetadata(ctx corectx.Context, cmd it.DeleteModelMetadataCommand) (*it.DeleteModelMetadataResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete model metadata", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *ModelMetadataDomainServiceImpl) ModelMetadataExists(ctx corectx.Context, query it.ModelMetadataExistsQuery) (*it.ModelMetadataExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if model metadata exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *ModelMetadataDomainServiceImpl) GetModelMetadata(ctx corectx.Context, query it.GetModelMetadataQuery) (*it.GetModelMetadataResult, error) {
	return corecrud.GetOne[models.ModelMetadata](ctx, corecrud.GetOneParam{
		Action: "get model metadata", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

func (this *ModelMetadataDomainServiceImpl) SearchModelMetadata(ctx corectx.Context, query it.SearchModelMetadataQuery) (*it.SearchModelMetadataResult, error) {
	return corecrud.Search[models.ModelMetadata](ctx, corecrud.SearchParam{
		Action: "search model metadata", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *ModelMetadataDomainServiceImpl) UpdateModelMetadata(ctx corectx.Context, cmd it.UpdateModelMetadataCommand) (*it.UpdateModelMetadataResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.ModelMetadata, *models.ModelMetadata]{
		Action: "update model metadata", DbRepoGetter: this.repo, Data: cmd,
	})
}
