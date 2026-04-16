package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

func NewModelMetadataServiceImpl(repo it.ModelMetadataRepository) it.ModelMetadataService {
	return &ModelMetadataServiceImpl{repo: repo}
}

type ModelMetadataServiceImpl struct{ repo it.ModelMetadataRepository }

func (this *ModelMetadataServiceImpl) CreateModelMetadata(ctx corectx.Context, cmd it.CreateModelMetadataCommand) (*it.CreateModelMetadataResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.ModelMetadata, *domain.ModelMetadata]{
		Action: "create model metadata", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *ModelMetadataServiceImpl) DeleteModelMetadata(ctx corectx.Context, cmd it.DeleteModelMetadataCommand) (*it.DeleteModelMetadataResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete model metadata", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *ModelMetadataServiceImpl) ModelMetadataExists(ctx corectx.Context, query it.ModelMetadataExistsQuery) (*it.ModelMetadataExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if model metadata exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *ModelMetadataServiceImpl) GetModelMetadata(ctx corectx.Context, query it.GetModelMetadataQuery) (*it.GetModelMetadataResult, error) {
	return corecrud.GetOne[domain.ModelMetadata](ctx, corecrud.GetOneParam{
		Action: "get model metadata", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

func (this *ModelMetadataServiceImpl) SearchModelMetadata(ctx corectx.Context, query it.SearchModelMetadataQuery) (*it.SearchModelMetadataResult, error) {
	return corecrud.Search[domain.ModelMetadata](ctx, corecrud.SearchParam{
		Action: "search model metadata", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *ModelMetadataServiceImpl) UpdateModelMetadata(ctx corectx.Context, cmd it.UpdateModelMetadataCommand) (*it.UpdateModelMetadataResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.ModelMetadata, *domain.ModelMetadata]{
		Action: "update model metadata", DbRepoGetter: this.repo, Data: cmd,
	})
}
