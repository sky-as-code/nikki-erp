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
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type OrganizationDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewOrganizationDynamicRepository(param OrganizationDynamicRepositoryParam) it.OrganizationRepository {
	dynamicRepo := baserepo.NewBaseDynamicRepository(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.OrganizationSchemaName),
		},
	)
	return &OrganizationDynamicRepository{dynamicRepo: dynamicRepo}
}

type OrganizationDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *OrganizationDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *OrganizationDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *OrganizationDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Organization) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *OrganizationDynamicRepository) Exists(ctx corectx.Context, keys []domain.Organization) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.Organization) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *OrganizationDynamicRepository) Insert(ctx corectx.Context, org domain.Organization) (
	*dyn.OpResult[int], error,
) {
	return baserepo.Insert(ctx, this.dynamicRepo, org)
}

func (this *OrganizationDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.Organization], error,
) {
	return baserepo.GetOne[domain.Organization](ctx, this.dynamicRepo, param)
}

func (this *OrganizationDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.Organization]], error) {
	return baserepo.Search[domain.Organization](ctx, this.dynamicRepo, param)
}

func (this *OrganizationDynamicRepository) Update(ctx corectx.Context, org domain.Organization) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, org.GetFieldData())
}
