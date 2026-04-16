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
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticket"
)

type TicketDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewTicketDynamicRepository(param TicketDynamicRepositoryParam) it.TicketRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.TicketSchemaName)})
	return &TicketDynamicRepository{dynamicRepo: dynamicRepo}
}

type TicketDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *TicketDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository { return this.dynamicRepo }
func (this *TicketDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *TicketDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Ticket) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *TicketDynamicRepository) Exists(ctx corectx.Context, keys []domain.Ticket) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Ticket) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *TicketDynamicRepository) Insert(ctx corectx.Context, data domain.Ticket) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *TicketDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Ticket], error) {
	return baserepo.GetOne[domain.Ticket](ctx, this.dynamicRepo, param)
}
func (this *TicketDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Ticket]], error) {
	return baserepo.Search[domain.Ticket](ctx, this.dynamicRepo, param)
}
func (this *TicketDynamicRepository) Update(ctx corectx.Context, data domain.Ticket) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
