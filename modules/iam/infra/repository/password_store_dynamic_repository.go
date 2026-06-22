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
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
)

type PasswordStoreDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewPasswordStoreDynamicRepository(param PasswordStoreDynamicRepositoryParam) it.PasswordStoreRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.PasswordStoreSchemaName),
		},
	)
	return &PasswordStoreDynamicRepository{dynamicRepo: dynamicRepo}
}

type PasswordStoreDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *PasswordStoreDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *PasswordStoreDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *PasswordStoreDynamicRepository) DeleteOne(
	ctx corectx.Context, keys models.PasswordStore,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *PasswordStoreDynamicRepository) Exists(
	ctx corectx.Context, keys []models.PasswordStore,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key models.PasswordStore) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *PasswordStoreDynamicRepository) Insert(
	ctx corectx.Context, passwordStore models.PasswordStore,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, passwordStore)
}

func (this *PasswordStoreDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[models.PasswordStore], error) {
	return baserepo.GetOne[models.PasswordStore](ctx, this.dynamicRepo, param)
}

func (this *PasswordStoreDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.PasswordStore]], error) {
	return baserepo.Search[models.PasswordStore](ctx, this.dynamicRepo, param)
}

func (this *PasswordStoreDynamicRepository) Update(
	ctx corectx.Context, passwordStore models.PasswordStore,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, passwordStore.GetFieldData())
}
