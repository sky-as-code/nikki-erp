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
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

type EscalationRuleDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewEscalationRuleDynamicRepository(param EscalationRuleDynamicRepositoryParam) it.EscalationRuleRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.EscalationRuleSchemaName)})
	return &EscalationRuleDynamicRepository{dynamicRepo: dynamicRepo}
}

type EscalationRuleDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *EscalationRuleDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *EscalationRuleDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *EscalationRuleDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.EscalationRule) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *EscalationRuleDynamicRepository) Exists(ctx corectx.Context, keys []domain.EscalationRule) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.EscalationRule) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *EscalationRuleDynamicRepository) Insert(ctx corectx.Context, data domain.EscalationRule) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, data)
}
func (this *EscalationRuleDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.EscalationRule], error) {
	return baserepo.GetOne[domain.EscalationRule](ctx, this.dynamicRepo, param)
}
func (this *EscalationRuleDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.EscalationRule]], error) {
	return baserepo.Search[domain.EscalationRule](ctx, this.dynamicRepo, param)
}
func (this *EscalationRuleDynamicRepository) Update(ctx corectx.Context, data domain.EscalationRule) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data.GetFieldData())
}
