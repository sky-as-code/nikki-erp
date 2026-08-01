package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// maxDescribedEntitlements bounds the single entitlement search behind a describe call. With
// at most DescribeRolesMaxIds roles per request this allows 100 entitlements per role, far
// above any realistic role, while still keeping the query bounded.
const maxDescribedEntitlements = itRole.DescribeRolesMaxIds * 100

// DescribeRoles resolves roles into the shape the role-assignment confirmation screen needs:
// every entitlement already carries its resource name, action name and scope name.
//
// The stored `expression` is deliberately not parsed. It is a `{action}:{resource}:{scope}`
// triple of opaque ids, and reading the structured columns instead keeps this endpoint correct
// regardless of what CalculateExpression writes into it.
func (this *RoleDomainServiceImpl) DescribeRoles(
	ctx corectx.Context, query itRole.DescribeRolesQuery,
) (result *itRole.DescribeRolesResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "describe roles"); e != nil {
			err = e
		}
	}()
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itRole.DescribeRolesResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itRole.DescribeRolesQuery))
	if len(query.RoleIds) == 0 {
		return &itRole.DescribeRolesResult{
			HasData: true,
			Data:    itRole.DescribeRolesResultData{Items: []itRole.DescribedRole{}},
		}, nil
	}

	roles, cErrs, err := this.describeRoleNames(ctx, query.RoleIds)
	if err != nil || cErrs.Count() > 0 {
		return &itRole.DescribeRolesResult{ClientErrors: cErrs}, err
	}
	byRole, cErrs, err := this.describeEntitlements(ctx, query.RoleIds)
	if err != nil || cErrs.Count() > 0 {
		return &itRole.DescribeRolesResult{ClientErrors: cErrs}, err
	}
	if err = this.fillScopeNames(ctx, byRole); err != nil {
		return nil, err
	}

	for i := range roles {
		roles[i].Entitlements = byRole[roles[i].Id]
		if roles[i].Entitlements == nil {
			roles[i].Entitlements = []itRole.DescribedEntitlement{}
		}
	}
	return &itRole.DescribeRolesResult{
		HasData: true,
		Data:    itRole.DescribeRolesResultData{Items: roles},
	}, nil
}

// describeRoleNames returns one entry per role that actually exists. Ids the caller cannot see
// are simply absent from the result rather than raising an error.
func (this *RoleDomainServiceImpl) describeRoleNames(
	ctx corectx.Context, roleIds []model.Id,
) ([]itRole.DescribedRole, ft.ClientErrors, error) {
	res, err := this.roleRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  idsInGraph(basemodel.FieldId, roleIds),
		Fields: []string{basemodel.FieldId, domain.RoleFieldName},
		Page:   0,
		Size:   len(roleIds),
	})
	if err != nil || res.ClientErrors.Count() > 0 {
		return nil, res.ClientErrors, err
	}
	roles := make([]itRole.DescribedRole, 0, len(res.Data.Items))
	for _, item := range res.Data.Items {
		fields := item.GetFieldData()
		id := fields.GetModelId(basemodel.FieldId)
		if id == nil {
			continue
		}
		roles = append(roles, itRole.DescribedRole{
			Id:   *id,
			Name: fields.GetString(domain.RoleFieldName),
		})
	}
	return roles, nil, nil
}

// describeEntitlements loads every entitlement of the given roles in one search, grouped by
// role id. `action.name` and `resource.name` are single-dot paths, which is the deepest select
// the query builder allows (MaxSelectGraphColumnDots), so both names arrive without extra
// round trips.
func (this *RoleDomainServiceImpl) describeEntitlements(
	ctx corectx.Context, roleIds []model.Id,
) (map[model.Id][]itRole.DescribedEntitlement, ft.ClientErrors, error) {
	res, err := this.entitlementRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: idsInGraph(domain.EntitlementFieldRoleId, roleIds),
		Fields: []string{
			basemodel.FieldId,
			domain.EntitlementFieldRoleId,
			domain.EntitlementFieldActionId,
			domain.EntitlementFieldResourceId,
			domain.EntitlementFieldScope,
			domain.EntitlementFieldOrgId,
			domain.EntitlementFieldOrgUnitId,
			domain.EntitlementEdgeAction + "." + domain.ActionFieldName,
			domain.EntitlementEdgeResource + "." + domain.ResourceFieldName,
		},
		Page: 0,
		Size: maxDescribedEntitlements,
	})
	if err != nil || res.ClientErrors.Count() > 0 {
		return nil, res.ClientErrors, err
	}

	byRole := make(map[model.Id][]itRole.DescribedEntitlement, len(roleIds))
	for _, item := range res.Data.Items {
		fields := item.GetFieldData()
		roleId := fields.GetModelId(domain.EntitlementFieldRoleId)
		id := fields.GetModelId(basemodel.FieldId)
		if roleId == nil || id == nil {
			continue
		}
		byRole[*roleId] = append(byRole[*roleId], describedEntitlementFrom(*id, fields))
	}
	return byRole, nil, nil
}

func describedEntitlementFrom(id model.Id, fields dmodel.DynamicFields) itRole.DescribedEntitlement {
	described := itRole.DescribedEntitlement{
		Id:           id,
		ResourceId:   fields.GetModelId(domain.EntitlementFieldResourceId),
		ResourceName: edgeName(fields, domain.EntitlementEdgeResource, domain.ResourceFieldName),
		ActionId:     fields.GetModelId(domain.EntitlementFieldActionId),
		ActionName:   edgeName(fields, domain.EntitlementEdgeAction, domain.ActionFieldName),
		Scope:        fields.GetString(domain.EntitlementFieldScope),
	}
	// Only org and orgunit entitlements target a specific record. Domain and private ones are
	// labelled by the caller from a static translation key, so they carry no id or name.
	if described.Scope == nil {
		return described
	}
	switch c.ResourceScope(*described.Scope) {
	case c.ResourceScopeOrg:
		described.ScopeId = fields.GetModelId(domain.EntitlementFieldOrgId)
	case c.ResourceScopeOrgUnit:
		described.ScopeId = fields.GetModelId(domain.EntitlementFieldOrgUnitId)
	}
	return described
}

// edgeName reads a leaf off a hydrated to-one edge. The repository stores the nested row under
// the bare edge key, not under the dotted path that was requested.
func edgeName(fields dmodel.DynamicFields, edge string, leaf string) *string {
	nested, ok := fields.GetAny(edge).(dmodel.DynamicFields)
	if !ok {
		return nil
	}
	return nested.GetString(leaf)
}

// fillScopeNames resolves the org and orgunit display names in one batched search each, rather
// than per entitlement.
func (this *RoleDomainServiceImpl) fillScopeNames(
	ctx corectx.Context, byRole map[model.Id][]itRole.DescribedEntitlement,
) error {
	orgIds, orgUnitIds := collectScopeIds(byRole)
	orgNames, err := this.searchNames(ctx, this.orgRepoSearch, orgIds, domain.OrgFieldDisplayName)
	if err != nil {
		return err
	}
	orgUnitNames, err := this.searchNames(ctx, this.orgUnitRepoSearch, orgUnitIds, domain.OrgUnitFieldName)
	if err != nil {
		return err
	}

	for roleId := range byRole {
		for i := range byRole[roleId] {
			entitlement := &byRole[roleId][i]
			if entitlement.ScopeId == nil || entitlement.Scope == nil {
				continue
			}
			switch c.ResourceScope(*entitlement.Scope) {
			case c.ResourceScopeOrg:
				entitlement.ScopeName = orgNames[*entitlement.ScopeId]
			case c.ResourceScopeOrgUnit:
				entitlement.ScopeName = orgUnitNames[*entitlement.ScopeId]
			}
		}
	}
	return nil
}

func collectScopeIds(byRole map[model.Id][]itRole.DescribedEntitlement) ([]model.Id, []model.Id) {
	orgs, orgUnits := map[model.Id]bool{}, map[model.Id]bool{}
	for roleId := range byRole {
		for _, entitlement := range byRole[roleId] {
			if entitlement.ScopeId == nil || entitlement.Scope == nil {
				continue
			}
			switch c.ResourceScope(*entitlement.Scope) {
			case c.ResourceScopeOrg:
				orgs[*entitlement.ScopeId] = true
			case c.ResourceScopeOrgUnit:
				orgUnits[*entitlement.ScopeId] = true
			}
		}
	}
	return keysOf(orgs), keysOf(orgUnits)
}

func keysOf(set map[model.Id]bool) []model.Id {
	ids := make([]model.Id, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	return ids
}

type namesSearchFn func(ctx corectx.Context, param dyn.RepoSearchParam) ([]dmodel.DynamicFields, ft.ClientErrors, error)

// fieldDataHaver is every dynamic-model entity; the describe flow only ever reads raw fields
// off a search result, so the concrete entity type is irrelevant here.
type fieldDataHaver interface {
	GetFieldData() dmodel.DynamicFields
}

func fieldRowsOf[T fieldDataHaver](items []T) []dmodel.DynamicFields {
	rows := make([]dmodel.DynamicFields, 0, len(items))
	for i := range items {
		rows = append(rows, items[i].GetFieldData())
	}
	return rows
}

func (this *RoleDomainServiceImpl) orgRepoSearch(
	ctx corectx.Context, param dyn.RepoSearchParam,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	res, err := this.orgRepo.Search(ctx, param)
	if err != nil || res.ClientErrors.Count() > 0 {
		return nil, res.ClientErrors, err
	}
	return fieldRowsOf(res.Data.Items), nil, nil
}

func (this *RoleDomainServiceImpl) orgUnitRepoSearch(
	ctx corectx.Context, param dyn.RepoSearchParam,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	res, err := this.orgUnitRepo.Search(ctx, param)
	if err != nil || res.ClientErrors.Count() > 0 {
		return nil, res.ClientErrors, err
	}
	return fieldRowsOf(res.Data.Items), nil, nil
}

func (this *RoleDomainServiceImpl) searchNames(
	ctx corectx.Context, search namesSearchFn, ids []model.Id, nameField string,
) (map[model.Id]*string, error) {
	names := map[model.Id]*string{}
	if len(ids) == 0 {
		return names, nil
	}
	rows, cErrs, err := search(ctx, dyn.RepoSearchParam{
		Graph:  idsInGraph(basemodel.FieldId, ids),
		Fields: []string{basemodel.FieldId, nameField},
		Page:   0,
		Size:   len(ids),
	})
	if err != nil {
		return nil, err
	}
	// A scope target the caller cannot read leaves the name nil; the label then falls back to
	// the raw scope on the frontend rather than failing the whole description.
	if cErrs.Count() > 0 {
		return names, nil
	}
	for _, row := range rows {
		id := row.GetModelId(basemodel.FieldId)
		if id != nil {
			names[*id] = row.GetString(nameField)
		}
	}
	return names, nil
}

func idsInGraph(field string, ids []model.Id) *dmodel.SearchGraph {
	ops := make([]any, len(ids))
	for i := range ids {
		ops[i] = ids[i]
	}
	graph := dmodel.NewSearchGraph()
	graph.Condition(dmodel.NewCondition(field, dmodel.In, ops...))
	return graph
}
