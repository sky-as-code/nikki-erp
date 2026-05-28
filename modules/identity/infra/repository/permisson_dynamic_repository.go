package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
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

func NewPermissionDynamicRepository(param UserPermissionRepositoryParam) itPerm.PermissionRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.UserPermissionSchemaName),
		},
	)
	return &PermissionDynamicRepository{
		dynamicRepo: dynamicRepo,
		orgUnitRepo: param.OrgUnitRepo,
	}
}

type PermissionDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
	orgUnitRepo itOrgUnit.OrgUnitRepository
}

func (this *PermissionDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *PermissionDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *PermissionDynamicRepository) RebuildUserPermission(ctx corectx.Context, userId model.Id) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", userId)
}

func (this *PermissionDynamicRepository) RebuildAllUserPermissions(ctx corectx.Context) error {
	return this.dynamicRepo.ExecFunc(ctx, "authz_rebuild_user_perm", nil)
}

func (this *PermissionDynamicRepository) MatchPermisions(ctx corectx.Context, param itPerm.RepoMatchUserPermParam) (*dyn.OpResult[[]models.UserPermission], error) {
	nodeUserId := dmodel.NewSearchNode().NewCondition(models.UserPermFieldUserId, dmodel.Equals, param.UserId)
	exactExpr := models.EntitlementExpression(&param.ActionCode, &param.ResourceCode, param.Scope, param.ScopeId)
	allActThisRsrcExpr := models.EntitlementExpression(nil, &param.ResourceCode, param.Scope, param.ScopeId)
	thisActAllRsrcExpr := models.EntitlementExpression(&param.ActionCode, nil, param.Scope, param.ScopeId)
	allActAllRsrcExpr := models.EntitlementExpression(nil, nil, param.Scope, param.ScopeId)

	exprNodes := []dmodel.SearchNode{
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, exactExpr),
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allActThisRsrcExpr),
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, thisActAllRsrcExpr),
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allActAllRsrcExpr),
	}

	if param.Scope == c.ResourceScopeOrg || param.Scope == c.ResourceScopeOrgUnit {
		exactDomainExpr := models.EntitlementExpression(&param.ActionCode, &param.ResourceCode, c.ResourceScopeDomain)
		allActDomainExpr := models.EntitlementExpression(nil, &param.ResourceCode, c.ResourceScopeDomain)
		allResDomainExpr := models.EntitlementExpression(nil, nil, c.ResourceScopeDomain)
		exprNodes = append(exprNodes,
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, exactDomainExpr),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allActDomainExpr),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allResDomainExpr),
		)
	}
	if param.Scope == c.ResourceScopeOrgUnit {
		exactOrgExpr := models.EntitlementExpression(&param.ActionCode, &param.ResourceCode, c.ResourceScopeOrg, param.ScopeId)
		allActOrgExpr := models.EntitlementExpression(nil, &param.ResourceCode, c.ResourceScopeOrg, param.ScopeId)
		allResOrgExpr := models.EntitlementExpression(nil, nil, c.ResourceScopeOrg, param.ScopeId)
		exprNodes = append(exprNodes,
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, exactOrgExpr),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allActOrgExpr),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allResOrgExpr),
		)

		resOrgUnit, err := this.orgUnitRepo.GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{
				models.OrgUnitFieldId: param.ScopeId,
			},
			Fields: []string{models.OrgUnitFieldOrgId},
		})
		if err != nil {
			return nil, err
		}
		if resOrgUnit.HasData {
			ordId := resOrgUnit.Data.GetOrgId()
			exactOrgIdExpr := models.EntitlementExpression(&param.ActionCode, &param.ResourceCode, c.ResourceScopeOrg, ordId)
			allActOrgIdExpr := models.EntitlementExpression(nil, &param.ResourceCode, c.ResourceScopeOrg, ordId)
			allResOrgIdExpr := models.EntitlementExpression(nil, nil, c.ResourceScopeOrg, ordId)
			exprNodes = append(exprNodes,
				*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, exactOrgIdExpr),
				*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allActOrgIdExpr),
				*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.Equals, allResOrgIdExpr),
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
		result := &dyn.OpResult[[]models.UserPermission]{
			HasData: true,
		}
		result.Data = array.Map(resSearch.Data.Items, func(item dmodel.DynamicFields) models.UserPermission {
			var userPermission models.UserPermission
			userPermission.SetFieldData(item)
			return userPermission
		})
		return result, nil
	}
	return &dyn.OpResult[[]models.UserPermission]{
		HasData: false,
	}, nil
}

// func (this *PermissionDynamicRepository) ListByUser(
// 	ctx corectx.Context, param itPerm.RepoListByUserParam,
// ) (*dyn.OpResult[[]models.UserPermission], error) {
// 	graph := &dmodel.SearchGraph{}
// 	if param.UserId != nil {
// 		graph.NewCondition(models.UserPermFieldUserId, dmodel.Equals, *param.UserId)
// 	}
// 	if param.UserEmail != nil {
// 		graph.NewCondition(fmt.Sprintf("%s.%s", models.UserPermEdgeUser, models.UserFieldEmail), dmodel.Equals, *param.UserEmail)
// 	}

// 	result, err := baserepo.Search[models.UserPermission](ctx, this.dynamicRepo, dyn.RepoSearchParam{
// 		Graph:  graph,
// 		Fields: param.Fields,
// 		Page:   0,
// 		Size:   math.MaxInt32,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if result.ClientErrors.Count() > 0 || !result.HasData {
// 		return &dyn.OpResult[[]models.UserPermission]{
// 			ClientErrors: result.ClientErrors,
// 			HasData:      result.HasData,
// 		}, nil
// 	}

// 	return &dyn.OpResult[[]models.UserPermission]{
// 		HasData: result.HasData,
// 		Data:    result.Data.Items,
// 	}, nil
// }

func (this *PermissionDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.UserPermission]], error) {
	return baserepo.Search[models.UserPermission](ctx, this.dynamicRepo, param)
}
