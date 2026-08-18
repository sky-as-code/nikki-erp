package repository

import (
	"fmt"
	"time"

	"github.com/lib/pq"
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
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
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
	return this.dynamicRepo.ExecFunc(ctx, "iam_rebuild_user_perm", userId)
}

func (this *PermissionDynamicRepository) RebuildAllUserPermissions(ctx corectx.Context) error {
	return this.dynamicRepo.ExecFunc(ctx, "iam_rebuild_all_user_perms")
}

// RebuildUserPermissionsForRole refreshes the cached rows of every holder of the
// role. One statement regardless of holder count, and it joins the ambient
// transaction so the mutation and its cache effect commit together.
func (this *PermissionDynamicRepository) RebuildUserPermissionsForRole(ctx corectx.Context, roleId model.Id) error {
	return this.dynamicRepo.ExecFunc(ctx, "iam_rebuild_perms_for_role", roleId)
}

// RebuildUserPermissionsForGroup rebuilds permissions of every member of the group in a single
// round-trip. It joins the ambient transaction when the context carries one.
func (this *PermissionDynamicRepository) RebuildUserPermissionsForGroup(ctx corectx.Context, groupId model.Id) error {
	relTable := dmodel.MustGetSchema(models.GrpUsrRelSchemaName).TableName()
	sqlQuery := fmt.Sprintf(
		`SELECT iam_rebuild_user_perm(%s) FROM %s WHERE %s = $1`,
		models.GrpUsrRelFieldUserId, relTable, models.GrpUsrRelFieldGroupId,
	)
	_, err := this.dynamicRepo.ExtractClient(ctx).Exec(ctx.InnerContext(), sqlQuery, groupId)
	return err
}

// FindMatchingPermissions returns the full provenance rows that answer the
// question, so the caller can report which grant paths are responsible.
//
// MatchPermisions answers yes/no from the same graph; this one keeps the rows.
func (this *PermissionDynamicRepository) FindMatchingPermissions(
	ctx corectx.Context, userId model.Id, candidates []string,
) ([]models.UserPermission, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	resSearch, err := this.dynamicRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: matchGraph(userId, candidates),
		Fields: []string{
			models.UserPermFieldEntId, models.UserPermFieldEntExpression,
			models.UserPermFieldSourceKind, models.UserPermFieldSourceId,
		},
	})
	if err != nil {
		return nil, err
	}
	if !resSearch.HasData {
		return nil, nil
	}
	return array.Map(resSearch.Data.Items, func(item dmodel.DynamicFields) models.UserPermission {
		var userPermission models.UserPermission
		userPermission.SetFieldData(item)
		return userPermission
	}), nil
}

// ResolveGrantSourceNames maps assignment ids to the name a person recognises -
// the role name for a direct assignment, the group name for a group assignment.
//
// Two statements at most, whatever the number of matches. Written as raw SQL
// because the assignment schemas declare no edge to the role or group, and adding
// one only to render a label would change the shape of every assignment query.
func (this *PermissionDynamicRepository) ResolveGrantSourceNames(
	ctx corectx.Context, directIds []model.Id, groupIds []model.Id,
) (map[model.Id]string, error) {
	names := make(map[model.Id]string, len(directIds)+len(groupIds))

	queries := []struct {
		ids   []model.Id
		query string
	}{
		{directIds, `SELECT ra.id, r.name FROM iam_role_user_assignments ra
			JOIN iam_roles r ON r.id = ra.role_id WHERE ra.id = ANY($1)`},
		{groupIds, `SELECT ra.id, g.name FROM iam_role_group_assignments ra
			JOIN iam_groups g ON g.id = ra.receiver_group_id WHERE ra.id = ANY($1)`},
	}
	for _, q := range queries {
		if len(q.ids) == 0 {
			continue
		}
		if err := this.collectNames(ctx, q.query, q.ids, names); err != nil {
			return nil, err
		}
	}
	return names, nil
}

func (this *PermissionDynamicRepository) collectNames(
	ctx corectx.Context, query string, ids []model.Id, names map[model.Id]string,
) error {
	rawIds := make([]string, len(ids))
	for i, id := range ids {
		rawIds[i] = string(id)
	}

	// pq.Array, not the bare slice: the driver cannot marshal a Go slice into the
	// array literal that `= ANY($1)` expects.
	rows, err := this.dynamicRepo.ExtractClient(ctx).Query(ctx.InnerContext(), query, pq.Array(rawIds))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		names[model.Id(id)] = name
	}
	return rows.Err()
}

// buildRequiredPerm turns the query into the Perm the shared candidate generator
// expects. For a unit-scoped question it also resolves the unit's org, which is
// what lets an org-level grant answer for a record inside that org's unit.
func (this *PermissionDynamicRepository) buildRequiredPerm(
	ctx corectx.Context, param itPerm.RepoMatchUserPermParam,
) (*reguard.Perm, error) {
	required := reguard.PermFor(param.ActionCode, param.ResourceCode, reguard.ResourceScope(param.Scope))
	required.IsRecordOwnedByCaller = param.IsRecordOwnedByCaller

	switch param.Scope {
	case c.ResourceScopeOrg:
		required = required.InOrg(param.ScopeId)

	case c.ResourceScopeOrgUnit:
		required = required.InOrgUnit(param.ScopeId)
		if param.ScopeId == nil {
			break
		}
		resOrgUnit, err := this.orgUnitRepo.GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.OrgUnitFieldId: *param.ScopeId},
			Fields: []string{models.OrgUnitFieldOrgId},
		})
		if err != nil {
			return nil, err
		}
		if resOrgUnit.HasData {
			required = required.InOrg(resOrgUnit.Data.GetOrgId())
		}
	}
	return &required, nil
}

// matchGraph is the one query shape that answers "does this user hold any of
// these expressions right now". Both the SQL matcher and the permission probe use
// it, so the probe cannot report a grant the matcher would not honour.
//
// It is a single range scan on (user_id, ent_expression), plus the expiry filter
// that makes an expired grant stop answering without waiting for a rebuild.
func matchGraph(userId model.Id, candidates []string) *dmodel.SearchGraph {
	graph := dmodel.NewSearchGraph()
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldUserId, dmodel.Equals, userId),
		*dmodel.NewSearchNode().NewCondition(models.UserPermFieldEntExpression, dmodel.In, toAnySlice(candidates)...),
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldExpiresAt, dmodel.IsNotSet),
			*dmodel.NewSearchNode().NewCondition(models.UserPermFieldExpiresAt, dmodel.GreaterThan, time.Now()),
		),
	)
	return graph
}

func toAnySlice(items []string) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}

// MatchPermisions returns the cached rows that answer the required permission.
//
// The candidate expressions come from requestguard.CandidateExpressions - the same
// function the in-memory guard and the permission probe use - so this path cannot
// answer a question differently from the way it is enforced at the middleware.
func (this *PermissionDynamicRepository) MatchPermisions(ctx corectx.Context, param itPerm.RepoMatchUserPermParam) (*dyn.OpResult[[]models.UserPermission], error) {
	required, err := this.buildRequiredPerm(ctx, param)
	if err != nil {
		return nil, err
	}
	candidates := reguard.CandidateExpressions(*required, param.EvalContext)

	resSearch, err := this.dynamicRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: matchGraph(param.UserId, candidates),
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
