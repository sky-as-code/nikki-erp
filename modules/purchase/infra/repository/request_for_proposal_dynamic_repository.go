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
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

type RequestForProposalDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewRequestForProposalDynamicRepository(param RequestForProposalDynamicRepositoryParam) it.RequestForProposalRepository {
	repo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{
		Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger,
		Schema: dmodel.MustGetSchema(domain.RequestForProposalSchemaName),
	})
	return &RequestForProposalDynamicRepository{dynamicRepo: repo}
}

type RequestForProposalDynamicRepository struct{ dynamicRepo dyn.BaseDynamicRepository }

func (this *RequestForProposalDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *RequestForProposalDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *RequestForProposalDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *RequestForProposalDynamicRepository) Exists(ctx corectx.Context, keys []domain.RequestForProposal) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, array.Map(keys, func(k domain.RequestForProposal) dmodel.DynamicFields { return k.GetFieldData() }))
}
func (this *RequestForProposalDynamicRepository) Insert(ctx corectx.Context, input domain.RequestForProposal) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, input)
}
func (this *RequestForProposalDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.RequestForProposal], error) {
	return baserepo.GetOne[domain.RequestForProposal](ctx, this.dynamicRepo, param)
}
func (this *RequestForProposalDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.RequestForProposal]], error) {
	return baserepo.Search[domain.RequestForProposal](ctx, this.dynamicRepo, param)
}
func (this *RequestForProposalDynamicRepository) Update(ctx corectx.Context, input domain.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, input.GetFieldData())
}
