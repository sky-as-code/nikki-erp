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
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

type SlaPolicyDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewSlaPolicyDynamicRepository(param SlaPolicyDynamicRepositoryParam) it.SlaPolicyRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.SlaPolicySchemaName)})
	return &SlaPolicyDynamicRepository{dynamicRepo: dynamicRepo}
}

type SlaPolicyDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *SlaPolicyDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *SlaPolicyDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *SlaPolicyDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.SlaPolicy) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *SlaPolicyDynamicRepository) Exists(ctx corectx.Context, keys []domain.SlaPolicy) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.SlaPolicy) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *SlaPolicyDynamicRepository) Insert(ctx corectx.Context, data domain.SlaPolicy) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *SlaPolicyDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.SlaPolicy], error) {
	return baserepo.GetOne[domain.SlaPolicy](ctx, this.dynamicRepo, param)
}
func (this *SlaPolicyDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.SlaPolicy]], error) {
	return baserepo.Search[domain.SlaPolicy](ctx, this.dynamicRepo, param)
}
func (this *SlaPolicyDynamicRepository) Update(ctx corectx.Context, data domain.SlaPolicy) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
