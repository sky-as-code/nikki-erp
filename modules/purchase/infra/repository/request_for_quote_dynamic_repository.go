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
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforquote"
)

type RequestForQuoteDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewRequestForQuoteDynamicRepository(param RequestForQuoteDynamicRepositoryParam) it.RequestForQuoteRepository {
	repo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{
		Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger,
		Schema: dmodel.MustGetSchema(domain.RequestForQuoteSchemaName),
	})
	return &RequestForQuoteDynamicRepository{dynamicRepo: repo}
}

type RequestForQuoteDynamicRepository struct{ dynamicRepo dyn.BaseDynamicRepository }

func (this *RequestForQuoteDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *RequestForQuoteDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *RequestForQuoteDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.RequestForQuote) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *RequestForQuoteDynamicRepository) Exists(ctx corectx.Context, keys []domain.RequestForQuote) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, array.Map(keys, func(k domain.RequestForQuote) dmodel.DynamicFields { return k.GetFieldData() }))
}
func (this *RequestForQuoteDynamicRepository) Insert(ctx corectx.Context, input domain.RequestForQuote) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, input)
}
func (this *RequestForQuoteDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.RequestForQuote], error) {
	return baserepo.GetOne[domain.RequestForQuote](ctx, this.dynamicRepo, param)
}
func (this *RequestForQuoteDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.RequestForQuote]], error) {
	return baserepo.Search[domain.RequestForQuote](ctx, this.dynamicRepo, param)
}
func (this *RequestForQuoteDynamicRepository) Update(ctx corectx.Context, input domain.RequestForQuote) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, input.GetFieldData())
}
