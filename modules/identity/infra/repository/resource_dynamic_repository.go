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
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

type ResourceDynamicRepositoryParam struct {
	dig.In

	Client        dyorm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  dyorm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewResourceDynamicRepository(param ResourceDynamicRepositoryParam) it.ResourceRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.ResourceSchemaName),
		},
	)
	return &ResourceDynamicRepository{
		dynamicRepo:  dynamicRepo,
		queryBuilder: param.QueryBuilder,
	}
}

type ResourceDynamicRepository struct {
	dynamicRepo  dyn.BaseDynamicRepository
	queryBuilder dyorm.QueryBuilder
}

func (this *ResourceDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *ResourceDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *ResourceDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Resource) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *ResourceDynamicRepository) Exists(ctx corectx.Context, keys []domain.Resource) (
	*dyn.OpResult[dyn.RepoExistsResult], error,
) {
	dynamicKeys := array.Map(keys, func(key domain.Resource) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *ResourceDynamicRepository) Insert(ctx corectx.Context, row domain.Resource) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *ResourceDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.Resource], error,
) {
	return baserepo.GetOne[domain.Resource](ctx, this.dynamicRepo, param)
}

func (this *ResourceDynamicRepository) GetByAction(ctx corectx.Context, param it.RepoGetByActionParam) (*dyn.OpResult[domain.Resource], error) {
	path := domain.ResourceEdgeActions + "." + domain.ActionFieldCode
	graph := dmodel.NewSearchGraph().NewCondition(path, dmodel.Equals, param.ActionCode)
	result, err := baserepo.Search[domain.Resource](ctx, this.dynamicRepo, dyn.RepoSearchParam{
		Graph:   graph,
		Columns: param.Columns,
	})
	if err != nil {
		return nil, err
	}
	if !result.HasData {
		return nil, nil
	}
	return &dyn.OpResult[domain.Resource]{
		Data:    result.Data.Items[0],
		HasData: result.HasData,
	}, nil
}

func (this *ResourceDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[domain.Resource]], error,
) {
	return baserepo.Search[domain.Resource](ctx, this.dynamicRepo, param)
}

func (this *ResourceDynamicRepository) Update(ctx corectx.Context, row domain.Resource) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}
