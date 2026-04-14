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
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

type ProductCategoryDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewProductCategoryDynamicRepository(param ProductCategoryDynamicRepositoryParam) it.ProductCategoryRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.ProductCategorySchemaName),
		},
	)
	return &ProductCategoryDynamicRepository{dynamicRepo: dynamicRepo}
}

type ProductCategoryDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *ProductCategoryDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *ProductCategoryDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *ProductCategoryDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.ProductCategory) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *ProductCategoryDynamicRepository) Exists(ctx corectx.Context, keys []domain.ProductCategory) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.ProductCategory) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *ProductCategoryDynamicRepository) Insert(
	ctx corectx.Context, productCategory domain.ProductCategory,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, productCategory)
}

func (this *ProductCategoryDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.ProductCategory], error) {
	return baserepo.GetOne[domain.ProductCategory](ctx, this.dynamicRepo, param)
}

func (this *ProductCategoryDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.ProductCategory]], error) {
	return baserepo.Search[domain.ProductCategory](ctx, this.dynamicRepo, param)
}

func (this *ProductCategoryDynamicRepository) Update(
	ctx corectx.Context, productCategory domain.ProductCategory,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, productCategory.GetFieldData())
}
