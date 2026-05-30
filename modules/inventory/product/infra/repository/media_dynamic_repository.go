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
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/media"
)

type InventoryMediaDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewInventoryMediaDynamicRepository(param InventoryMediaDynamicRepositoryParam) it.InventoryMediaRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.InventoryMediaSchemaName),
		},
	)
	return &InventoryMediaDynamicRepository{dynamicRepo: dynamicRepo}
}

type InventoryMediaDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *InventoryMediaDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *InventoryMediaDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *InventoryMediaDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.InventoryMedia) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *InventoryMediaDynamicRepository) Exists(ctx corectx.Context, keys []domain.InventoryMedia) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.InventoryMedia) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *InventoryMediaDynamicRepository) Insert(
	ctx corectx.Context, product domain.InventoryMedia,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, product)
}

func (this *InventoryMediaDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.InventoryMedia], error) {
	return baserepo.GetOne[domain.InventoryMedia](ctx, this.dynamicRepo, param)
}

func (this *InventoryMediaDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.InventoryMedia]], error) {
	return baserepo.Search[domain.InventoryMedia](ctx, this.dynamicRepo, param)
}

func (this *InventoryMediaDynamicRepository) Update(
	ctx corectx.Context, product domain.InventoryMedia,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, product.GetFieldData())
}
