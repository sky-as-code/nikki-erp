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
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

type OrgUnitDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewOrgUnitDynamicRepository(param OrgUnitDynamicRepositoryParam) it.OrgUnitRepository {
	dynamicRepo := baserepo.NewBaseDynamicRepository(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.OrganizationalUnitSchemaName),
		},
	)
	return &OrgUnitDynamicRepository{dynamicRepo: dynamicRepo}
}

type OrgUnitDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *OrgUnitDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *OrgUnitDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *OrgUnitDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.OrganizationalUnit) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *OrgUnitDynamicRepository) Exists(ctx corectx.Context, keys []domain.OrganizationalUnit) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.OrganizationalUnit) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *OrgUnitDynamicRepository) Insert(
	ctx corectx.Context, level domain.OrganizationalUnit,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, level)
}

func (this *OrgUnitDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.OrganizationalUnit], error) {
	return baserepo.GetOne[domain.OrganizationalUnit](ctx, this.dynamicRepo, param)
}

func (this *OrgUnitDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.OrganizationalUnit]], error) {
	return baserepo.Search[domain.OrganizationalUnit](ctx, this.dynamicRepo, param)
}

func (this *OrgUnitDynamicRepository) Update(ctx corectx.Context, level domain.OrganizationalUnit) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, level.GetFieldData())
}
