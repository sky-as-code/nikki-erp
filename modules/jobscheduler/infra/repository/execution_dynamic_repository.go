package repository

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"

	domain "github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/execution"
)

func NewExecutionDynamicRepository(param JobDynamicRepositoryParam) it.ExecutionRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.ExecutionSchemaName),
		},
	)
	return &ExecutionDynamicRepository{dynamicRepo: dynamicRepo}
}

type ExecutionDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *ExecutionDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *ExecutionDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *ExecutionDynamicRepository) Insert(
	ctx corectx.Context, execution domain.Execution,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, execution)
}

func (this *ExecutionDynamicRepository) Update(
	ctx corectx.Context, execution domain.Execution,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, execution.GetFieldData())
}

func (this *ExecutionDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.Execution], error) {
	return baserepo.GetOne[domain.Execution](ctx, this.dynamicRepo, param)
}

func (this *ExecutionDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.Execution]], error) {
	return baserepo.Search[domain.Execution](ctx, this.dynamicRepo, param)
}
