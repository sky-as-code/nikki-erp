package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
)

type UserPermissionRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
	OrgUnitRepo   itOrgUnit.OrgUnitRepository
}

func NewUserPermissionRepositoryImpl(param UserPermissionRepositoryParam) itPerm.UserPermissionRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.UserPermissionSchemaName),
		},
	)
	return &UserPermissionRepositoryImpl{
		dynamicRepo: dynamicRepo,
		orgUnitRepo: param.OrgUnitRepo,
	}
}

type UserPermissionRepositoryImpl struct {
	dynamicRepo dyn.BaseDynamicRepository
	orgUnitRepo itOrgUnit.OrgUnitRepository
}

func (this *UserPermissionRepositoryImpl) RebuildUserPermission(ctx corectx.Context, userId model.Id) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", userId)
}

func (this *UserPermissionRepositoryImpl) RebuildAllUserPermissions(ctx corectx.Context) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", nil)
}

func (this *UserPermissionRepositoryImpl) MatchPermisions(ctx corectx.Context, param itPerm.RepoMatchUserPermParam) (*dyn.OpResult[[]domain.UserPermission], error) {
	nodeUserId := dmodel.NewSearchNode().NewCondition(domain.UserPermFieldUserId, dmodel.Equals, param.UserId)
	exactExpr := domain.EntitlementExpression(&param.ActionCode, &param.ResourceCode, param.Scope, param.ScopeId)
	allActExpr := domain.EntitlementExpression(nil, &param.ResourceCode, param.Scope, param.ScopeId)
	allResExpr := domain.EntitlementExpression(nil, nil, param.Scope, param.ScopeId)

	exprNodes := []dmodel.SearchNode{
		*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, exactExpr),
		*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allActExpr),
		*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allResExpr),
	}

	if param.Scope == domain.ResourceScopeOrg || param.Scope == domain.ResourceScopeOrgUnit {
		exactDomainExpr := domain.EntitlementExpression(&param.ActionCode, &param.ResourceCode, domain.ResourceScopeDomain)
		allActDomainExpr := domain.EntitlementExpression(nil, &param.ResourceCode, domain.ResourceScopeDomain)
		allResDomainExpr := domain.EntitlementExpression(nil, nil, domain.ResourceScopeDomain)
		exprNodes = append(exprNodes,
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, exactDomainExpr),
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allActDomainExpr),
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allResDomainExpr),
		)
	}
	if param.Scope == domain.ResourceScopeOrgUnit {
		exactOrgExpr := domain.EntitlementExpression(&param.ActionCode, &param.ResourceCode, domain.ResourceScopeOrg, param.ScopeId)
		allActOrgExpr := domain.EntitlementExpression(nil, &param.ResourceCode, domain.ResourceScopeOrg, param.ScopeId)
		allResOrgExpr := domain.EntitlementExpression(nil, nil, domain.ResourceScopeOrg, param.ScopeId)
		exprNodes = append(exprNodes,
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, exactOrgExpr),
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allActOrgExpr),
			*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allResOrgExpr),
		)

		resOrgUnit, err := this.orgUnitRepo.GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{
				domain.OrgUnitFieldId: param.ScopeId,
			},
			Columns: []string{domain.OrgUnitFieldOrgId},
		})
		if err != nil {
			return nil, err
		}
		if resOrgUnit.HasData {
			ordId := resOrgUnit.Data.GetOrgId()
			exactOrgIdExpr := domain.EntitlementExpression(&param.ActionCode, &param.ResourceCode, domain.ResourceScopeOrg, ordId)
			allActOrgIdExpr := domain.EntitlementExpression(nil, &param.ResourceCode, domain.ResourceScopeOrg, ordId)
			allResOrgIdExpr := domain.EntitlementExpression(nil, nil, domain.ResourceScopeOrg, ordId)
			exprNodes = append(exprNodes,
				*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, exactOrgIdExpr),
				*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allActOrgIdExpr),
				*dmodel.NewSearchNode().NewCondition(domain.UserPermFieldEntExpression, dmodel.Equals, allResOrgIdExpr),
			)
		}
	}
	graph := dmodel.NewSearchGraph()
	graph.And(
		*nodeUserId,
		*dmodel.NewSearchNode().Or(exprNodes...),
	)
	resSearch, err := this.dynamicRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
	})
	if err != nil {
		return nil, err
	}
	if resSearch.HasData {
		result := &dyn.OpResult[[]domain.UserPermission]{
			HasData: true,
		}
		result.Data = array.Map(resSearch.Data.Items, func(item dmodel.DynamicFields) domain.UserPermission {
			var userPermission domain.UserPermission
			userPermission.SetFieldData(item)
			return userPermission
		})
		return result, nil
	}
	return &dyn.OpResult[[]domain.UserPermission]{
		HasData: false,
	}, nil
}
