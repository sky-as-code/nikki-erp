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
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

type VariantDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewVariantDynamicRepository(param VariantDynamicRepositoryParam) it.VariantRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.VariantSchemaName),
		},
	)
	return &VariantDynamicRepository{dynamicRepo: dynamicRepo}
}

type VariantDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *VariantDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *VariantDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *VariantDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Variant) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *VariantDynamicRepository) Exists(ctx corectx.Context, keys []domain.Variant) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Variant) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *VariantDynamicRepository) Insert(
	ctx corectx.Context, variant domain.Variant,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, variant)
}

func (this *VariantDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.Variant], error) {
	return baserepo.GetOne[domain.Variant](ctx, this.dynamicRepo, param)
}

func (this *VariantDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.Variant]], error) {
	return baserepo.Search[domain.Variant](ctx, this.dynamicRepo, param)
}

func (this *VariantDynamicRepository) Update(
	ctx corectx.Context, variant domain.Variant,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, variant.GetFieldData())
}
