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
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

type TicketActivityDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewTicketActivityDynamicRepository(param TicketActivityDynamicRepositoryParam) it.TicketActivityRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.TicketActivitySchemaName)})
	return &TicketActivityDynamicRepository{dynamicRepo: dynamicRepo}
}

type TicketActivityDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *TicketActivityDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *TicketActivityDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *TicketActivityDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.TicketActivity) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *TicketActivityDynamicRepository) Exists(ctx corectx.Context, keys []domain.TicketActivity) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.TicketActivity) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *TicketActivityDynamicRepository) Insert(ctx corectx.Context, data domain.TicketActivity) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *TicketActivityDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.TicketActivity], error) {
	return baserepo.GetOne[domain.TicketActivity](ctx, this.dynamicRepo, param)
}
func (this *TicketActivityDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.TicketActivity]], error) {
	return baserepo.Search[domain.TicketActivity](ctx, this.dynamicRepo, param)
}
func (this *TicketActivityDynamicRepository) Update(ctx corectx.Context, data domain.TicketActivity) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
