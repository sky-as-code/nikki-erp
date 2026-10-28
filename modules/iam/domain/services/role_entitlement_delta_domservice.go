package services

import (
	"fmt"
	"sort"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

const (
	deltaFieldAdd    = itRole.UpdateRoleFieldEntitlements + ".add"
	deltaFieldRemove = itRole.UpdateRoleFieldEntitlements + ".remove"
	deltaFieldUpdate = itRole.UpdateRoleFieldEntitlements + ".update"
)

// roleGrantPlan is one requested grant, resolved and checked before anything is written.
type roleGrantPlan struct {
	field       string
	resourceId  model.Id
	actionId    *model.Id
	description *string
	scope       c.ResourceScope
	orgId       *model.Id
	expression  string
}

// applyEntitlementsDelta expects the caller's transaction: the removals, the grants, their audit
// rows and the single permission rebuild must commit or roll back together. Removals run first
// so that a delta may revoke and re-grant the same expression. Description updates run last:
// they touch only rows the role keeps. The grants added and removed are also recorded on the role,
// whose history would otherwise never show them.
func (this *RoleCrudDomainServiceImpl) applyEntitlementsDelta(
	ctx corectx.Context, roleId model.Id, delta itRole.RoleEntitlementsDelta,
) (ft.ClientErrors, error) {
	if cErrs := validateDeltaSize(delta); cErrs.Count() > 0 {
		return cErrs, nil
	}
	if cErrs := validateEntitlementUpdates(delta); cErrs.Count() > 0 {
		return cErrs, nil
	}
	removed, cErrs, err := this.removeRoleEntitlements(ctx, roleId, delta.Remove)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	added, cErrs, err := this.addRoleEntitlements(ctx, roleId, delta.Add)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	cErrs, err = this.updateRoleEntitlements(ctx, roleId, delta.Update)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	cErrs, err = this.recordEntitlementLinks(ctx, roleId, added, removed)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	return nil, this.permRepo.RebuildUserPermissionsForRole(ctx, roleId)
}

// recordEntitlementLinks records expressions rather than ids: a removed entitlement row is
// deleted, so the history could no longer name it by id.
func (this *RoleCrudDomainServiceImpl) recordEntitlementLinks(
	ctx corectx.Context, roleId model.Id, added, removed []string,
) (ft.ClientErrors, error) {
	if len(added) == 0 && len(removed) == 0 {
		return nil, nil
	}
	orgId, cErrs, err := this.roleOrgIdOf(ctx, roleId)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	return nil, composable.RecordLinkAudit(ctx, this.roleRepo.Schema(), roleId, orgId,
		domain.RoleEdgeEntitlements, added, removed)
}

func validateDeltaSize(delta itRole.RoleEntitlementsDelta) ft.ClientErrors {
	var out ft.ClientErrors
	maxItems := itRole.RoleEntitlementsDeltaMaxItems
	vars := map[string]any{"max": maxItems}
	if len(delta.Add) > maxItems {
		out = append(out, *ft.NewValidationError(deltaFieldAdd, ft.ErrorKey("err_too_many_items", "iam"),
			"at most {{max}} entitlements can be added in one request", vars))
	}
	if len(delta.Remove) > maxItems {
		out = append(out, *ft.NewValidationError(deltaFieldRemove, ft.ErrorKey("err_too_many_items", "iam"),
			"at most {{max}} entitlements can be removed in one request", vars))
	}
	if len(delta.Update) > maxItems {
		out = append(out, *ft.NewValidationError(deltaFieldUpdate, ft.ErrorKey("err_too_many_items", "iam"),
			"at most {{max}} entitlements can be updated in one request", vars))
	}
	return out
}

// removeRoleEntitlements returns the expressions it removed.
func (this *RoleCrudDomainServiceImpl) removeRoleEntitlements(
	ctx corectx.Context, roleId model.Id, ids []model.Id,
) ([]string, ft.ClientErrors, error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}
	owned, cErrs, err := this.searchRoleEntitlements(ctx, roleId, basemodel.FieldId, idsToAny(ids))
	if err != nil || cErrs.Count() > 0 {
		return nil, prefixFields(cErrs, deltaFieldRemove), err
	}
	var out ft.ClientErrors
	for i, entId := range ids {
		if _, isOwned := owned[string(entId)]; !isOwned {
			out = append(out, *ft.NewBusinessViolation(
				fmt.Sprintf("%s[%d]", deltaFieldRemove, i), ft.ErrorKey("err_entitlement_not_in_role", "iam"),
				"entitlement {{id}} does not belong to this role", map[string]any{"id": string(entId)},
			))
		}
	}
	if out.Count() > 0 {
		return nil, out, nil
	}
	removed := make([]string, 0, len(owned))
	for entId, expression := range owned {
		res, err := this.entitlementSvc.DeleteDeferRebuild(ctx, dmodel.DynamicFields{basemodel.FieldId: entId})
		if err != nil {
			return nil, nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return nil, prefixFields(res.ClientErrors, deltaFieldRemove), nil
		}
		removed = append(removed, expression)
	}
	sort.Strings(removed)
	return removed, nil, nil
}

// addRoleEntitlements returns the expressions it granted.
func (this *RoleCrudDomainServiceImpl) addRoleEntitlements(
	ctx corectx.Context, roleId model.Id, grants []itRole.RoleEntitlementGrant,
) ([]string, ft.ClientErrors, error) {
	if len(grants) == 0 {
		return nil, nil, nil
	}
	roleOrgId, cErrs, err := this.roleOrgIdOf(ctx, roleId)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	plans, cErrs, err := this.planRoleGrants(ctx, roleOrgId, grants)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	cErrs, err = this.assertGrantsNotHeld(ctx, roleId, plans)
	if err != nil || cErrs.Count() > 0 {
		return nil, cErrs, err
	}
	added := make([]string, 0, len(plans))
	for _, plan := range plans {
		res, err := this.entitlementSvc.CreateDeferRebuild(ctx, plan.command(roleId))
		if err != nil {
			return nil, nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return nil, prefixFields(res.ClientErrors, plan.field), nil
		}
		added = append(added, plan.expression)
	}
	return added, nil, nil
}

func (this *RoleCrudDomainServiceImpl) roleOrgIdOf(
	ctx corectx.Context, roleId model.Id,
) (*model.Id, ft.ClientErrors, error) {
	roleRes, err := corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "read role org for entitlement grants",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.GetOneQuery{Id: roleId, Fields: []string{domain.RoleFieldOrgId}},
	})
	if err != nil {
		return nil, nil, err
	}
	if roleRes.ClientErrors.Count() > 0 {
		return nil, roleRes.ClientErrors, nil
	}
	if !roleRes.HasData {
		return nil, ft.ClientErrors{*ft.NewNotFoundError(basemodel.FieldId)}, nil
	}
	orgId := roleRes.Data.GetFieldData().GetModelId(domain.RoleFieldOrgId)
	if orgId != nil && *orgId == "" {
		orgId = nil
	}
	return orgId, nil, nil
}

// planRoleGrants resolves every grant and refuses a delta that asks for the same expression
// twice: the second insert would hit the (role_id, expression) unique key as a server error.
func (this *RoleCrudDomainServiceImpl) planRoleGrants(
	ctx corectx.Context, roleOrgId *model.Id, grants []itRole.RoleEntitlementGrant,
) ([]roleGrantPlan, ft.ClientErrors, error) {
	resources := make(map[model.Id]*domain.Resource)
	firstIndex := make(map[string]int)
	plans := make([]roleGrantPlan, 0, len(grants))
	var out ft.ClientErrors
	for i, grant := range grants {
		field := fmt.Sprintf("%s[%d]", deltaFieldAdd, i)
		plan, cErr, err := this.planRoleGrant(ctx, resources, roleOrgId, grant, field)
		if err != nil {
			return nil, nil, err
		}
		if cErr != nil {
			out = append(out, *cErr)
			continue
		}
		if prev, isDup := firstIndex[plan.expression]; isDup {
			out = append(out, *ft.NewBusinessViolation(field, ft.ErrorKey("err_entitlement_duplicated", "iam"),
				"the same grant is requested twice: {{expression}} (also at index {{index}})",
				map[string]any{"expression": plan.expression, "index": prev}))
			continue
		}
		firstIndex[plan.expression] = i
		plans = append(plans, *plan)
	}
	return plans, out, nil
}

func (this *RoleCrudDomainServiceImpl) planRoleGrant(
	ctx corectx.Context, resources map[model.Id]*domain.Resource, roleOrgId *model.Id,
	grant itRole.RoleEntitlementGrant, field string,
) (*roleGrantPlan, *ft.ClientErrorItem, error) {
	if grant.ResourceId == "" {
		return nil, dmodel.NewMissingFieldErr(field + "." + domain.EntitlementFieldResourceId), nil
	}
	resource, cErr, err := this.grantResource(ctx, resources, grant.ResourceId, field)
	if err != nil || cErr != nil {
		return nil, cErr, err
	}
	plan, cErr := planGrantOnResource(*resource, roleOrgId, grant, field)
	return plan, cErr, nil
}

// planGrantOnResource is the pure part of planning: it checks the action belongs to the
// resource, derives the scope from max_scope and computes the expression the entitlement
// service will store, so duplicates can be found before the insert.
func planGrantOnResource(
	resource domain.Resource, roleOrgId *model.Id, grant itRole.RoleEntitlementGrant, field string,
) (*roleGrantPlan, *ft.ClientErrorItem) {
	actionCode, cErr := grantActionCode(resource, grant.ActionId, field)
	if cErr != nil {
		return nil, cErr
	}
	scope, orgId, cErr := grantScope(resource, roleOrgId, field)
	if cErr != nil {
		return nil, cErr
	}
	return &roleGrantPlan{
		field:       field,
		resourceId:  grant.ResourceId,
		actionId:    grant.ActionId,
		description: grant.Description,
		scope:       scope,
		orgId:       orgId,
		expression:  domain.EntitlementExpression(actionCode, resource.GetCode(), scope, orgId),
	}, nil
}

func (this *RoleCrudDomainServiceImpl) grantResource(
	ctx corectx.Context, resources map[model.Id]*domain.Resource, resourceId model.Id, field string,
) (*domain.Resource, *ft.ClientErrorItem, error) {
	if cached, found := resources[resourceId]; found {
		return cached, nil, nil
	}
	resRes, err := this.resourceRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{domain.ResourceFieldId: resourceId},
		Fields: []string{
			domain.ResourceFieldCode,
			domain.ResourceFieldMaxScope,
			fmt.Sprintf("%s.%s", domain.ResourceEdgeActions, domain.ActionFieldId),
			fmt.Sprintf("%s.%s", domain.ResourceEdgeActions, domain.ActionFieldCode),
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if resRes.ClientErrors.Count() > 0 || !resRes.HasData {
		return nil, ft.NewNotFoundError(field + "." + domain.EntitlementFieldResourceId), nil
	}
	resource := *domain.NewResourceFrom(resRes.Data)
	resources[resourceId] = &resource
	return &resource, nil, nil
}

// grantActionCode returns nil for a nil action id, which the expression renders as `*`.
func grantActionCode(
	resource domain.Resource, actionId *model.Id, field string,
) (*string, *ft.ClientErrorItem) {
	if actionId == nil {
		return nil, nil
	}
	for _, action := range resource.GetActions() {
		if id := action.GetId(); id != nil && *id == *actionId {
			return action.GetCode(), nil
		}
	}
	return nil, ft.NewNotFoundError(field + "." + domain.EntitlementFieldActionId)
}

// grantScope gives the grant its resource's max_scope. Only tenant and org can be granted from
// a role: org_unit and private need a target the role does not carry.
func grantScope(
	resource domain.Resource, roleOrgId *model.Id, field string,
) (c.ResourceScope, *model.Id, *ft.ClientErrorItem) {
	maxScope := resource.GetMaxScope()
	if maxScope == nil {
		return "", nil, ft.NewNotFoundError(field + "." + domain.EntitlementFieldResourceId)
	}
	switch *maxScope {
	case c.ResourceScopeTenant:
		return *maxScope, nil, nil
	case c.ResourceScopeOrg:
		if roleOrgId == nil {
			return "", nil, ft.NewBusinessViolation(field, ft.ErrorKey("err_role_org_required", "iam"),
				"the resource is org-scoped, so only a role that belongs to an org can grant it")
		}
		return *maxScope, roleOrgId, nil
	default:
		return "", nil, ft.NewBusinessViolation(field, ft.ErrorKey("err_scope_not_grantable", "iam"),
			"a resource whose max_scope is {{scope}} cannot be granted from a role",
			map[string]any{"scope": string(*maxScope)})
	}
}

func (this *RoleCrudDomainServiceImpl) assertGrantsNotHeld(
	ctx corectx.Context, roleId model.Id, plans []roleGrantPlan,
) (ft.ClientErrors, error) {
	expressions := make([]any, len(plans))
	for i := range plans {
		expressions[i] = plans[i].expression
	}
	held, cErrs, err := this.searchRoleEntitlements(ctx, roleId, domain.EntitlementFieldExpression, expressions)
	if err != nil || cErrs.Count() > 0 {
		return prefixFields(cErrs, deltaFieldAdd), err
	}
	var out ft.ClientErrors
	for _, plan := range plans {
		if _, isHeld := held[plan.expression]; isHeld {
			out = append(out, *ft.NewBusinessViolation(plan.field, ft.ErrorKey("err_entitlement_already_granted", "iam"),
				"the role already grants {{expression}}", map[string]any{"expression": plan.expression}))
		}
	}
	return out, nil
}

// searchRoleEntitlements returns the expressions of the role's entitlements whose keyField is one
// of values, keyed by that field. Archived rows are included: they still occupy the unique key.
func (this *RoleCrudDomainServiceImpl) searchRoleEntitlements(
	ctx corectx.Context, roleId model.Id, keyField string, values []any,
) (map[string]string, ft.ClientErrors, error) {
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(domain.EntitlementFieldRoleId, dmodel.Equals, roleId)),
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(keyField, dmodel.In, values...)),
	)
	searchRes, err := this.entitlementRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: []string{basemodel.FieldId, keyField, domain.EntitlementFieldExpression},
		Page:   0,
		Size:   len(values),
	})
	if err != nil {
		return nil, nil, err
	}
	if searchRes.ClientErrors.Count() > 0 {
		return nil, searchRes.ClientErrors, nil
	}
	found := make(map[string]string, len(searchRes.Data.Items))
	for _, fields := range searchRes.Data.Items {
		// fmt.Sprint covers both shapes a key may come back in: string and model.Id.
		if key, hasKey := fields[keyField]; hasKey && key != nil {
			found[fmt.Sprint(key)] = fmt.Sprint(fields[domain.EntitlementFieldExpression])
		}
	}
	return found, nil, nil
}

func (this roleGrantPlan) command(roleId model.Id) itEnt.CreateEntitlementCommand {
	ent := domain.NewEntitlement()
	name := this.expression
	ent.GetFieldData().SetString(domain.EntitlementFieldName, &name)
	if this.description != nil {
		ent.GetFieldData().SetString(domain.EntitlementFieldDescription, this.description)
	}
	scope := this.scope
	ent.SetRoleId(&roleId)
	ent.SetResourceId(&this.resourceId)
	ent.SetScope(&scope)
	if this.actionId != nil {
		ent.SetActionId(this.actionId)
	}
	if this.orgId != nil {
		ent.SetOrgId(this.orgId)
	}
	return ent.GetFieldData()
}

func idsToAny(ids []model.Id) []any {
	out := make([]any, len(ids))
	for i := range ids {
		out[i] = ids[i]
	}
	return out
}

// prefixFields re-homes errors raised against an entitlement field under the delta item that
// caused them, so the client can tell which row failed.
func prefixFields(cErrs ft.ClientErrors, prefix string) ft.ClientErrors {
	for i := range cErrs {
		if cErrs[i].Field == "" {
			cErrs[i].Field = prefix
			continue
		}
		cErrs[i].Field = prefix + "." + cErrs[i].Field
	}
	return cErrs
}
