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
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type GroupDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewGroupDynamicRepository(param GroupDynamicRepositoryParam) it.GroupRepository {
	dynamicRepo := baserepo.NewBaseDynamicRepository(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.GroupSchemaName),
		},
	)
	return &GroupDynamicRepository{dynamicRepo: dynamicRepo}
}

type GroupDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *GroupDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *GroupDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *GroupDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Group) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *GroupDynamicRepository) Exists(ctx corectx.Context, keys []domain.Group) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Group) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *GroupDynamicRepository) Insert(ctx corectx.Context, group domain.Group) (
	*dyn.OpResult[int], error,
) {
	return baserepo.Insert(ctx, this.dynamicRepo, group)
}

func (this *GroupDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.Group], error,
) {
	return baserepo.GetOne[domain.Group](ctx, this.dynamicRepo, param)
}

func (this *GroupDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.Group]], error) {
	return baserepo.Search[domain.Group](ctx, this.dynamicRepo, param)
}

func (this *GroupDynamicRepository) Update(ctx corectx.Context, group domain.Group) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, group.GetFieldData())
}
