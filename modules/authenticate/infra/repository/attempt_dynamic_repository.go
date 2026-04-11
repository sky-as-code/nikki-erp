package repository

import (
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type AttemptDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewAttemptDynamicRepository(param AttemptDynamicRepositoryParam) it.AttemptRepository {
	schema := dmodel.MustGetSchema(domain.LoginAttemptSchemaName)
	dynamicRepo := baserepo.NewBaseDynamicRepository(baserepo.NewBaseRepositoryParam{
		Client:       param.Client,
		ConfigSvc:    param.ConfigSvc,
		QueryBuilder: param.QueryBuilder,
		Logger:       param.Logger,
		Schema:       schema,
	})
	return &AttemptDynamicRepository{dynamicRepo: dynamicRepo}
}

type AttemptDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *AttemptDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *AttemptDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *AttemptDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.LoginAttempt], error) {
	return baserepo.GetOne[domain.LoginAttempt](ctx, this.dynamicRepo, param)
}

func (this *AttemptDynamicRepository) Insert(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, attempt)
}

func (this *AttemptDynamicRepository) Update(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, attempt.GetFieldData())
}
