package repository

import (
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

type HierarchyDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewHierarchyDynamicRepository(param HierarchyDynamicRepositoryParam) it.HierarchyRepository {
	dynamicRepo := baserepo.NewBaseRepositoryImpl(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.HierarchyLevelSchemaName),
		},
	)
	return &HierarchyDynamicRepository{dynamicRepo: dynamicRepo}
}

type HierarchyDynamicRepository struct {
	dynamicRepo dyn.BaseRepository
}

func (this *HierarchyDynamicRepository) GetBaseRepo() dyn.BaseRepository {
	return this.dynamicRepo
}

func (this *HierarchyDynamicRepository) Insert(
	ctx corectx.Context, level domain.HierarchyLevel,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, level)
}

func (this *HierarchyDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.HierarchyLevel], error) {
	return baserepo.GetOne[domain.HierarchyLevel](ctx, this.dynamicRepo, param)
}

func (this *HierarchyDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.HierarchyLevel]], error) {
	return baserepo.Search[domain.HierarchyLevel](ctx, this.dynamicRepo, param)
}

func (this *HierarchyDynamicRepository) Update(ctx corectx.Context, level domain.HierarchyLevel) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, level.GetFieldData())
}
