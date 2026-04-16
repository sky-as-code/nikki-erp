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
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/team"
)

type TeamDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewTeamDynamicRepository(param TeamDynamicRepositoryParam) it.TeamRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.TeamSchemaName)})
	return &TeamDynamicRepository{dynamicRepo: dynamicRepo}
}

type TeamDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *TeamDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository { return this.dynamicRepo }
func (this *TeamDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *TeamDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Team) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *TeamDynamicRepository) Exists(ctx corectx.Context, keys []domain.Team) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Team) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *TeamDynamicRepository) Insert(ctx corectx.Context, data domain.Team) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *TeamDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Team], error) {
	return baserepo.GetOne[domain.Team](ctx, this.dynamicRepo, param)
}
func (this *TeamDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Team]], error) {
	return baserepo.Search[domain.Team](ctx, this.dynamicRepo, param)
}
func (this *TeamDynamicRepository) Update(ctx corectx.Context, data domain.Team) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
