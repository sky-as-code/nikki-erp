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
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

type UnitCategoryDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewUnitCategoryDynamicRepository(param UnitCategoryDynamicRepositoryParam) it.UnitCategoryRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.UnitCategorySchemaName),
		},
	)
	return &UnitCategoryDynamicRepository{dynamicRepo: dynamicRepo}
}

type UnitCategoryDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *UnitCategoryDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *UnitCategoryDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *UnitCategoryDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.UnitCategory) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *UnitCategoryDynamicRepository) Exists(ctx corectx.Context, keys []domain.UnitCategory) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.UnitCategory) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *UnitCategoryDynamicRepository) Insert(
	ctx corectx.Context, unitCategory domain.UnitCategory,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, unitCategory)
}

func (this *UnitCategoryDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.UnitCategory], error) {
	return baserepo.GetOne[domain.UnitCategory](ctx, this.dynamicRepo, param)
}

func (this *UnitCategoryDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.UnitCategory]], error) {
	return baserepo.Search[domain.UnitCategory](ctx, this.dynamicRepo, param)
}

func (this *UnitCategoryDynamicRepository) Update(ctx corectx.Context, unitCategory domain.UnitCategory) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, unitCategory.GetFieldData())
}
