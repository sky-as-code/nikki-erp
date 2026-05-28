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
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
)

type EntitlementDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewEntitlementDynamicRepository(param EntitlementDynamicRepositoryParam) it.EntitlementRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.EntitlementSchemaName),
		},
	)
	return &EntitlementDynamicRepository{dynamicRepo: dynamicRepo}
}

type EntitlementDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *EntitlementDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *EntitlementDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *EntitlementDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Entitlement) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *EntitlementDynamicRepository) Exists(ctx corectx.Context, keys []domain.Entitlement) (
	*dyn.OpResult[dyn.RepoExistsResult], error,
) {
	dynamicKeys := array.Map(keys, func(key domain.Entitlement) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *EntitlementDynamicRepository) Insert(ctx corectx.Context, row domain.Entitlement) (
	*dyn.OpResult[int], error,
) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *EntitlementDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.Entitlement], error,
) {
	return baserepo.GetOne[domain.Entitlement](ctx, this.dynamicRepo, param)
}

func (this *EntitlementDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[domain.Entitlement]], error,
) {
	return baserepo.Search[domain.Entitlement](ctx, this.dynamicRepo, param)
}

func (this *EntitlementDynamicRepository) Update(ctx corectx.Context, row domain.Entitlement) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}
