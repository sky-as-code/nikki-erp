package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyorm "github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
)

type RoleDynamicRepositoryParam struct {
	dig.In

	Client       dyorm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder dyorm.QueryBuilder
	Logger       logging.LoggerService
}

func NewRoleDynamicRepository(param RoleDynamicRepositoryParam) it.RoleRepository {
	dynamicRepo := baserepo.NewBaseDynamicRepository(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.RoleSchemaName),
		},
	)
	return &RoleDynamicRepository{dynamicRepo: dynamicRepo}
}

type RoleDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *RoleDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *RoleDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *RoleDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.Role) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *RoleDynamicRepository) Exists(ctx corectx.Context, keys []domain.Role) (
	*dyn.OpResult[dyn.RepoExistsResult], error,
) {
	dynamicKeys := array.Map(keys, func(key domain.Role) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *RoleDynamicRepository) Insert(ctx corectx.Context, row domain.Role) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, row)
}

func (this *RoleDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[domain.Role], error,
) {
	return baserepo.GetOne[domain.Role](ctx, this.dynamicRepo, param)
}

func (this *RoleDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[domain.Role]], error,
) {
	return baserepo.Search[domain.Role](ctx, this.dynamicRepo, param)
}

func (this *RoleDynamicRepository) Update(ctx corectx.Context, row domain.Role) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, row.GetFieldData())
}

func (this *RoleDynamicRepository) HasAssignedUsers(ctx corectx.Context, roleId model.Id) bool {
	return this.roleAssignmentCountPositive(ctx, domain.RoleEdgeAssignedUsers, roleId)
}

func (this *RoleDynamicRepository) HasAssignedGroups(ctx corectx.Context, roleId model.Id) bool {
	return this.roleAssignmentCountPositive(ctx, domain.RoleEdgeAssignedGroups, roleId)
}

func (this *RoleDynamicRepository) roleAssignmentCountPositive(
	ctx corectx.Context, m2mEdge string, roleId model.Id,
) bool {
	ok, err := this.dynamicRepo.ExistsM2m(ctx, dyn.RepoExistsM2mParam{
		M2mEdge: m2mEdge,
		SrcId:   roleId,
	})
	return err == nil && ok
}
