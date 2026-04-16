package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

func NewFieldMetadataServiceImpl(repo it.FieldMetadataRepository) it.FieldMetadataService {
	return &FieldMetadataServiceImpl{repo: repo}
}

type FieldMetadataServiceImpl struct{ repo it.FieldMetadataRepository }

func (this *FieldMetadataServiceImpl) CreateFieldMetadata(ctx corectx.Context, cmd it.CreateFieldMetadataCommand) (*it.CreateFieldMetadataResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.FieldMetadata, *domain.FieldMetadata]{
		Action: "create field metadata", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *FieldMetadataServiceImpl) DeleteFieldMetadata(ctx corectx.Context, cmd it.DeleteFieldMetadataCommand) (*it.DeleteFieldMetadataResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete field metadata", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *FieldMetadataServiceImpl) FieldMetadataExists(ctx corectx.Context, query it.FieldMetadataExistsQuery) (*it.FieldMetadataExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if field metadata exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *FieldMetadataServiceImpl) GetFieldMetadata(ctx corectx.Context, query it.GetFieldMetadataQuery) (*it.GetFieldMetadataResult, error) {
	return corecrud.GetOne[domain.FieldMetadata](ctx, corecrud.GetOneParam{
		Action: "get field metadata", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

func (this *FieldMetadataServiceImpl) SearchFieldMetadata(ctx corectx.Context, query it.SearchFieldMetadataQuery) (*it.SearchFieldMetadataResult, error) {
	return corecrud.Search[domain.FieldMetadata](ctx, corecrud.SearchParam{
		Action: "search field metadata", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *FieldMetadataServiceImpl) UpdateFieldMetadata(ctx corectx.Context, cmd it.UpdateFieldMetadataCommand) (*it.UpdateFieldMetadataResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.FieldMetadata, *domain.FieldMetadata]{
		Action: "update field metadata", DbRepoGetter: this.repo, Data: cmd,
	})
}
