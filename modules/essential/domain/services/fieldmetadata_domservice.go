package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

func NewFieldMetadataDomainServiceImpl(repo it.FieldMetadataRepository) it.FieldMetadataDomainService {
	return &FieldMetadataDomainServiceImpl{repo: repo}
}

type FieldMetadataDomainServiceImpl struct{ repo it.FieldMetadataRepository }

func (this *FieldMetadataDomainServiceImpl) CreateFieldMetadata(ctx corectx.Context, cmd it.CreateFieldMetadataCommand) (*it.CreateFieldMetadataResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.FieldMetadata, *models.FieldMetadata]{
		Action: "create field metadata", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *FieldMetadataDomainServiceImpl) DeleteFieldMetadata(ctx corectx.Context, cmd it.DeleteFieldMetadataCommand) (*it.DeleteFieldMetadataResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete field metadata", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *FieldMetadataDomainServiceImpl) FieldMetadataExists(ctx corectx.Context, query it.FieldMetadataExistsQuery) (*it.FieldMetadataExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if field metadata exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *FieldMetadataDomainServiceImpl) GetFieldMetadata(ctx corectx.Context, query it.GetFieldMetadataQuery) (*it.GetFieldMetadataResult, error) {
	return corecrud.GetOne[models.FieldMetadata](ctx, corecrud.GetOneParam{
		Action: "get field metadata", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

func (this *FieldMetadataDomainServiceImpl) SearchFieldMetadata(ctx corectx.Context, query it.SearchFieldMetadataQuery) (*it.SearchFieldMetadataResult, error) {
	return corecrud.Search[models.FieldMetadata](ctx, corecrud.SearchParam{
		Action: "search field metadata", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *FieldMetadataDomainServiceImpl) UpdateFieldMetadata(ctx corectx.Context, cmd it.UpdateFieldMetadataCommand) (*it.UpdateFieldMetadataResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.FieldMetadata, *models.FieldMetadata]{
		Action: "update field metadata", DbRepoGetter: this.repo, Data: cmd,
	})
}
