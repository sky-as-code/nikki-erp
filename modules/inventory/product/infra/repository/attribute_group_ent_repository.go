package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

type AttributeGroupDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewAttributeGroupDynamicRepository(param AttributeGroupDynamicRepositoryParam) it.AttributeGroupRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.AttributeGroupSchemaName),
		},
	)
	return &AttributeGroupDynamicRepository{dynamicRepo: dynamicRepo}
}

type AttributeGroupDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *AttributeGroupDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *AttributeGroupDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *AttributeGroupDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.AttributeGroup) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *AttributeGroupDynamicRepository) Exists(ctx corectx.Context, keys []domain.AttributeGroup) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.AttributeGroup) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *AttributeGroupDynamicRepository) Insert(
	ctx corectx.Context, attributeGroup domain.AttributeGroup,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, attributeGroup)
}

func (this *AttributeGroupDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.AttributeGroup], error) {
	return baserepo.GetOne[domain.AttributeGroup](ctx, this.dynamicRepo, param)
}

func (this *AttributeGroupDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.AttributeGroup]], error) {
	return baserepo.Search[domain.AttributeGroup](ctx, this.dynamicRepo, param)
}

func (this *AttributeGroupDynamicRepository) Update(
	ctx corectx.Context, attributeGroup domain.AttributeGroup,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, attributeGroup.GetFieldData())
}
