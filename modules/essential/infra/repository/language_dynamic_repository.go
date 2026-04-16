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
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

type LanguageDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewLanguageDynamicRepository(param LanguageDynamicRepositoryParam) it.LanguageRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.LanguageSchemaName),
		},
	)
	return &LanguageDynamicRepository{dynamicRepo: dynamicRepo}
}

type LanguageDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *LanguageDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *LanguageDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *LanguageDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Language) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *LanguageDynamicRepository) Exists(ctx corectx.Context, keys []domain.Language) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Language) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *LanguageDynamicRepository) Insert(ctx corectx.Context, src domain.Language) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, src)
}

func (this *LanguageDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Language], error) {
	return baserepo.GetOne[domain.Language](ctx, this.dynamicRepo, param)
}

func (this *LanguageDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Language]], error) {
	return baserepo.Search[domain.Language](ctx, this.dynamicRepo, param)
}

func (this *LanguageDynamicRepository) Update(ctx corectx.Context, src domain.Language) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, src.GetFieldData())
}
