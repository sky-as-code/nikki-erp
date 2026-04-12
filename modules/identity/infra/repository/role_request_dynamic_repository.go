package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyorm "github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

type RoleRequestDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewRoleRequestDynamicRepository(param RoleRequestDynamicRepositoryParam) it.RoleRequestRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.RoleRequestSchemaName),
		},
	)
	return &RoleRequestDynamicRepository{dynamicRepo: dynamicRepo}
}

type RoleRequestDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *RoleRequestDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *RoleRequestDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *RoleRequestDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.RoleRequest) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *RoleRequestDynamicRepository) Exists(ctx corectx.Context, keys []domain.RoleRequest) (
	*dyn.OpResult[dyn.RepoExistsResult], error,
) {
	dynamicKeys := array.Map(keys, func(key domain.RoleRequest) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *RoleRequestDynamicRepository) Insert(ctx corectx.Context, row domain.RoleRequest) (
	*dyn.OpResult[int], error,
) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *RoleRequestDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.RoleRequest], error,
) {
	return baserepo.GetOne[domain.RoleRequest](ctx, this.dynamicRepo, param)
}

func (this *RoleRequestDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[domain.RoleRequest]], error,
) {
	return baserepo.Search[domain.RoleRequest](ctx, this.dynamicRepo, param)
}

func (this *RoleRequestDynamicRepository) Update(ctx corectx.Context, row domain.RoleRequest) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}
