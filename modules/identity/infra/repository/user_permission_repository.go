package repository

import (
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
)

type UserPermissionRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewUserPermissionRepositoryImpl(param UserPermissionRepositoryParam) it.UserPermissionRepository {
	dynamicRepo := baserepo.NewBaseDynamicRepository(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.UserPermissionSchemaName),
		},
	)
	return &UserPermissionRepositoryImpl{dynamicRepo: dynamicRepo}
}

type UserPermissionRepositoryImpl struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *UserPermissionRepositoryImpl) RebuildUserPermission(ctx corectx.Context, userId model.Id) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", userId)
}

func (this *UserPermissionRepositoryImpl) RebuildAllUserPermissions(ctx corectx.Context) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", nil)
}

func (this *UserPermissionRepositoryImpl) GetOne(ctx corectx.Context, param it.GetUserPermissionParam) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.dynamicRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			domain.UserPermFieldUserId:       param.UserId,
			domain.UserPermFieldActionCode:   param.ActionCode,
			domain.UserPermFieldResourceCode: param.ResourceCode,
			domain.UserPermFieldScope:        param.Scope,
			domain.UserPermFieldOrgId:        param.OrgId,
			domain.UserPermFieldOrgUnitId:    param.OrgUnitId,
		},
	})
}
