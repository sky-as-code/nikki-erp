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
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketassignment"
)

type TicketAssignmentDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewTicketAssignmentDynamicRepository(param TicketAssignmentDynamicRepositoryParam) it.TicketAssignmentRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(models.TicketAssignmentSchemaName)})
	return &TicketAssignmentDynamicRepository{dynamicRepo: dynamicRepo}
}

type TicketAssignmentDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *TicketAssignmentDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *TicketAssignmentDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *TicketAssignmentDynamicRepository) DeleteOne(ctx corectx.Context, keys models.TicketAssignment) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *TicketAssignmentDynamicRepository) Exists(ctx corectx.Context, keys []models.TicketAssignment) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key models.TicketAssignment) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *TicketAssignmentDynamicRepository) Insert(ctx corectx.Context, data models.TicketAssignment) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *TicketAssignmentDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.TicketAssignment], error) {
	return baserepo.GetOne[models.TicketAssignment](ctx, this.dynamicRepo, param)
}
func (this *TicketAssignmentDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.TicketAssignment]], error) {
	return baserepo.Search[models.TicketAssignment](ctx, this.dynamicRepo, param)
}
func (this *TicketAssignmentDynamicRepository) Update(ctx corectx.Context, data models.TicketAssignment) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
