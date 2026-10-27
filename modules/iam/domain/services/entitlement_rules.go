package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// The validation and derivation rules of an entitlement: its scope must fit the resource's
// bounds and name an existing org or unit, and its expression is built from the codes behind
// the action and resource ids.

func (this *EntitlementCrudDomainServiceImpl) roleIdOfEntitlement(
	ctx corectx.Context, entitlementId model.Id,
) (*model.Id, error) {
	resEnt, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{domain.EntitlementFieldId: entitlementId},
		Fields: []string{domain.EntitlementFieldRoleId},
	})
	if err != nil {
		return nil, err
	}
	if !resEnt.HasData {
		return nil, nil
	}
	return resEnt.Data.GetModelId(domain.EntitlementFieldRoleId), nil
}

// calculateExpression resolves the action and resource **codes** behind the ids the
// caller sent, then lets the entitlement build its expression from them.
//
// The runtime guard looks for expressions built from codes, so resolving them here
// is what makes an API-created entitlement enforceable at all (plan defect D1). A
// nil id means "all actions" / "all resources" and becomes the wildcard, so a
// missing id is not an error.
func (this *EntitlementCrudDomainServiceImpl) calculateExpression(
	ctx corectx.Context, ent *domain.Entitlement, vErrs *ft.ClientErrors,
) error {
	actionCode, err := this.resolveActionCode(ctx, ent.GetActionId(), vErrs)
	if err != nil {
		return err
	}
	resourceCode, err := this.resolveResourceCode(ctx, ent.GetResourceId(), vErrs)
	if err != nil {
		return err
	}
	if vErrs.Count() > 0 {
		return nil
	}

	ent.CalculateExpression(actionCode, resourceCode)
	applyFieldName(ent, actionCode, resourceCode)
	return nil
}

// applyFieldName narrows the expression of a field-level entitlement to its field:
// `{action}:{resource}.{field_name}:{scope}`. The guard matches exactly what
// reguard.BuildFieldExpression produces, so the expression is rebuilt through it rather than
// patched as a string. A resource-level entitlement (no field_name) keeps what CalculateExpression
// wrote; resource_id is required, so a field never hangs off the wildcard resource.
func applyFieldName(ent *domain.Entitlement, actionCode *string, resourceCode *string) {
	fieldName := ent.GetFieldData().GetString(domain.EntitlementFieldFieldName)
	scope := ent.GetScope()
	if fieldName == nil || *fieldName == "" || scope == nil || resourceCode == nil {
		return
	}
	scopeId := ent.GetOrgId()
	if *scope == c.ResourceScopeOrgUnit {
		scopeId = ent.GetOrgUnitId()
	}
	expr := reguard.BuildFieldExpression(
		valueOr(actionCode, reguard.Wildcard), *resourceCode, *fieldName, *scope, scopeId,
	)
	ent.GetFieldData().SetString(domain.EntitlementFieldExpression, &expr)
}

func valueOr(val *string, fallback string) string {
	if val == nil {
		return fallback
	}
	return *val
}

func (this *EntitlementCrudDomainServiceImpl) resolveActionCode(
	ctx corectx.Context, actionId *model.Id, vErrs *ft.ClientErrors,
) (*string, error) {
	if actionId == nil {
		return nil, nil
	}
	resAction, err := this.actionRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{domain.ActionFieldId: *actionId},
		Fields: []string{domain.ActionFieldCode},
	})
	if err != nil {
		return nil, err
	}
	if !resAction.HasData {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldActionId))
		return nil, nil
	}
	return domain.NewActionFrom(resAction.Data).GetCode(), nil
}

func (this *EntitlementCrudDomainServiceImpl) resolveResourceCode(
	ctx corectx.Context, resourceId *model.Id, vErrs *ft.ClientErrors,
) (*string, error) {
	if resourceId == nil {
		return nil, nil
	}
	resResource, err := this.resourceRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{domain.ResourceFieldId: *resourceId},
		Fields: []string{domain.ResourceFieldCode},
	})
	if err != nil {
		return nil, err
	}
	if !resResource.HasData {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldResourceId))
		return nil, nil
	}
	return domain.NewResourceFrom(resResource.Data).GetCode(), nil
}

func (this *EntitlementCrudDomainServiceImpl) validateScope(
	ctx corectx.Context, ent *domain.Entitlement, cErrsTotal *ft.ClientErrors,
) error {
	resourceId := ent.GetResourceId()
	if resourceId == nil {
		cErrsTotal.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldResourceId))
		return nil
	}
	resource, err := this.fetchResourceForAction(ctx, resourceId, cErrsTotal)
	if err != nil {
		return err
	}
	if cErrsTotal.Count() > 0 || resource == nil {
		return nil
	}
	// A resource whose scope bounds or an entitlement whose scope is absent is a
	// client error, not a crash: every one of these was a nil dereference reported
	// to the caller as a 500.
	minScopePtr, maxScopePtr, entScopePtr := resource.GetMinScope(), resource.GetMaxScope(), ent.GetScope()
	if minScopePtr == nil || maxScopePtr == nil {
		cErrsTotal.Append(*ft.NewNotFoundError(domain.EntitlementFieldResourceId))
		return nil
	}
	if entScopePtr == nil {
		cErrsTotal.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldScope))
		return nil
	}

	minScope := *minScopePtr
	maxScope := *maxScopePtr
	entScope := *entScopePtr
	orgId := ent.GetOrgId()
	orgUnitId := ent.GetOrgUnitId()

	cErrs, err := dyn.StartValidationFlowCopy(cErrsTotal).
		Step(func(cErrs *ft.ClientErrors) error {
			return this.checkOrgExistence(ctx, entScope, orgId, cErrs)
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			return this.checkOrgUnitExistence(ctx, entScope, orgUnitId, cErrs)
		}).
		Step(func(cErrs *ft.ClientErrors) error {
			if !domain.IsResourceScopeInBounds(minScope, maxScope, entScope) {
				cErrs.Append(*ft.NewValidationError(
					domain.EntitlementFieldScope, ft.ErrorKey("err_scope_out_of_bounds", "iam"),
					"scope must be between the resource min_scope and max_scope (both inclusive)",
				))
			}
			return nil
		}).
		End()
	cErrsTotal.Concat(cErrs)

	return err
}

func (this *EntitlementCrudDomainServiceImpl) checkOrgExistence(
	ctx corectx.Context, entScope c.ResourceScope, orgId *model.Id, vErrs *ft.ClientErrors,
) error {
	if entScope == c.ResourceScopeOrg && orgId == nil {
		vErrs.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldOrgId))
		return nil
	}
	if entScope == c.ResourceScopeOrg && orgId != nil {
		existence, err := this.orgRepo.KeyExist(ctx, dmodel.DynamicFields{basemodel.FieldId: string(*orgId)})
		if err != nil {
			return err
		}
		if existence.ClientErrors.Count() > 0 || !existence.Data {
			vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldOrgId))
			return nil
		}
	}
	return nil
}

func (this *EntitlementCrudDomainServiceImpl) checkOrgUnitExistence(
	ctx corectx.Context, entScope c.ResourceScope, orgUnitId *model.Id, vErrs *ft.ClientErrors,
) error {
	if entScope == c.ResourceScopeOrgUnit && orgUnitId == nil {
		vErrs.Append(*dmodel.NewMissingFieldErr(domain.EntitlementFieldOrgUnitId))
		return nil
	}
	if entScope == c.ResourceScopeOrgUnit && orgUnitId != nil {
		existence, err := this.orgUnitRepo.KeyExist(ctx, dmodel.DynamicFields{basemodel.FieldId: string(*orgUnitId)})
		if err != nil {
			return err
		}
		if existence.ClientErrors.Count() > 0 || !existence.Data {
			vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldOrgUnitId))
			return nil
		}
	}
	return nil
}

func (this *EntitlementCrudDomainServiceImpl) fetchResourceForAction(
	ctx corectx.Context, resourceId *model.Id, vErrs *ft.ClientErrors,
) (*domain.Resource, error) {

	resRes, err := this.resourceRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			domain.ResourceFieldId: *resourceId,
		},
		// min_scope and max_scope are what the caller actually validates against;
		// selecting only the actions edge left them nil, so the scope-bounds check
		// had nothing to compare and every create failed.
		Fields: []string{
			domain.ResourceFieldMinScope,
			domain.ResourceFieldMaxScope,
			fmt.Sprintf("%s.%s", domain.ResourceEdgeActions, domain.ActionFieldId),
		},
	})
	if err != nil {
		return nil, err
	}
	if !resRes.HasData {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldResourceId))
		return nil, nil
	}
	r := *domain.NewResourceFrom(resRes.Data)
	if r.GetActions() == nil {
		vErrs.Append(*ft.NewNotFoundError(domain.EntitlementFieldActionId))
		return nil, nil
	}
	return &r, nil
}
