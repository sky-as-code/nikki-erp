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

func NewAttemptDynamicRepository(param JobDynamicRepositoryParam) it.AttemptRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.AttemptSchemaName),
		},
	)
	return &AttemptDynamicRepository{dynamicRepo: dynamicRepo}
}

type AttemptDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *AttemptDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *AttemptDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *AttemptDynamicRepository) Insert(
	ctx corectx.Context, attempt domain.Attempt,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, attempt)
}

func (this *AttemptDynamicRepository) Update(
	ctx corectx.Context, attempt domain.Attempt,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, attempt.GetFieldData())
}

func (this *AttemptDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.Attempt], error) {
	return baserepo.GetOne[domain.Attempt](ctx, this.dynamicRepo, param)
}

func (this *AttemptDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.Attempt]], error) {
	return baserepo.Search[domain.Attempt](ctx, this.dynamicRepo, param)
}
