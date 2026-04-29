package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyorm "github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

type UserPreferenceDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewUserPreferenceDynamicRepository(param UserPreferenceDynamicRepositoryParam) it.UserPreferenceRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.UserPreferenceSchemaName),
		},
	)
	return &UserPreferenceDynamicRepository{
		dynamicRepo:  dynamicRepo,
		queryBuilder: param.QueryBuilder,
	}
}

type UserPreferenceDynamicRepository struct {
	dynamicRepo  dyn.BaseDynamicRepository
	queryBuilder dyorm.QueryBuilder
}

func (this *UserPreferenceDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *UserPreferenceDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *UserPreferenceDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *UserPreferenceDynamicRepository) Exists(ctx corectx.Context, keys []domain.UserPreference) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.UserPreference) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *UserPreferenceDynamicRepository) Insert(ctx corectx.Context, row domain.UserPreference) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *UserPreferenceDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.UserPreference], error) {
	return baserepo.GetOne[domain.UserPreference](ctx, this.dynamicRepo, param)
}

func (this *UserPreferenceDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.UserPreference]], error) {
	return baserepo.Search[domain.UserPreference](ctx, this.dynamicRepo, param)
}

func (this *UserPreferenceDynamicRepository) Update(ctx corectx.Context, row domain.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}
