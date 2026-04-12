package repository

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
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
	schema := dmodel.MustGetSchema(domain.PasswordStoreSchemaName)
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       schema,
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

func (this *PasswordStoreDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.PasswordStore) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *PasswordStoreDynamicRepository) Insert(ctx corectx.Context, pass domain.PasswordStore) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, pass)
}

func (this *PasswordStoreDynamicRepository) Update(
	ctx corectx.Context, pass domain.PasswordStore,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, pass.GetFieldData())
}

func (this *PasswordStoreDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.PasswordStore], error) {
	return baserepo.GetOne[domain.PasswordStore](ctx, this.dynamicRepo, param)
}
