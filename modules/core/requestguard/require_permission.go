package requestguard

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type Perm struct {
	ResourceCode string
	ActionCode   string
	Scope        ResourceScope
	// This is Org Unit ID to which the resource belongs (if any).
	// If this is not nil, then OrgId must be this Org Unit's Org ID.
	OrgUnitId *model.Id

	// This is the Org ID to which the resource belongs (if any)
	// Or, this can be the Org Unit's Org ID (if the resource belongs to an org unit)
	OrgId *model.Id

	// IsRecordOwnedByCaller answers the private scope: the caller may act on their
	// own record. Callers whose scope is private must set it; leaving it false
	// simply means "not the caller's record", which denies.
	IsRecordOwnedByCaller bool
}

// PermFor starts a Perm for the given action on the given resource. Use the
// InOrg / InOrgUnit / OwnedByCaller builders to attach the record's context - an
// org- or unit-scoped check without it can only ever match an exact or domain
// grant, which is how org-scoped checks silently degraded before.
func PermFor(actionCode string, resourceCode string, scope ResourceScope) Perm {
	return Perm{ActionCode: actionCode, ResourceCode: resourceCode, Scope: scope}
}

// InOrg names the org the record belongs to.
func (this Perm) InOrg(orgId *model.Id) Perm {
	this.OrgId = orgId
	return this
}

// InOrgUnit names the org unit the record belongs to. Pass the unit's org as well
// via InOrg so that the org-level fallback can apply.
func (this Perm) InOrgUnit(orgUnitId *model.Id) Perm {
	this.OrgUnitId = orgUnitId
	return this
}

// OwnedByCaller marks the record as the caller's own, which is what a private
// scope grant is about.
func (this Perm) OwnedByCaller(owned bool) Perm {
	this.IsRecordOwnedByCaller = owned
	return this
}

type PermissionContext struct {
	IsOwner      bool
	Entitlements []string
}

// AssertPermission answers whether the caller may perform requiredPerm.
//
// It holds no rules of its own: it asks CandidateExpressions which stored
// expressions would answer, and checks whether the caller holds any of them. The
// SQL matcher and the permission probe ask the same function, which is what keeps
// the three of them from drifting into disagreeing about the same question.
func AssertPermission(ctx corectx.Context, requiredPerm Perm) *ft.ClientErrors {
	userPerm := ctx.GetPermissions()
	if userPerm.IsOwner {
		return nil
	}

	// The private scope is about the record, not about the expression: a private
	// grant answers only for the caller's own record.
	if requiredPerm.Scope == ResourceScopePrivate && !requiredPerm.IsRecordOwnedByCaller {
		return insufficient(requiredPerm)
	}

	evalCtx := EvalContextFrom(userPerm)
	for _, candidate := range CandidateExpressions(requiredPerm, evalCtx) {
		if userPerm.Entitlements.Contains(candidate) {
			return nil
		}
	}

	return insufficient(requiredPerm)
}

// EvalContextFrom lifts the caller's org and unit membership out of the request
// context, so callers of CandidateExpressions do not each reach into it.
func EvalContextFrom(userPerm corectx.ContextPermissions) EvalContext {
	return EvalContext{
		UserOrgIds:   userPerm.UserOrgIds.ToSlice(),
		OrgUnitId:    userPerm.OrgUnitId,
		OrgUnitOrgId: userPerm.OrgUnitOrgId,
	}
}

// insufficient names the exact expression the caller was missing, which is the
// one piece of information that makes a 403 actionable for an administrator.
func insufficient(requiredPerm Perm) *ft.ClientErrors {
	exact := BuildExpression(
		requiredPerm.ActionCode, requiredPerm.ResourceCode,
		requiredPerm.Scope, scopeIdOf(requiredPerm),
	)
	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewInsufficientPermissionsError([]string{exact}))
	return cErrs
}

func scopeIdOf(requiredPerm Perm) *model.Id {
	if requiredPerm.Scope == ResourceScopeOrgUnit {
		return requiredPerm.OrgUnitId
	}
	if requiredPerm.Scope == ResourceScopeOrg {
		return requiredPerm.OrgId
	}
	return nil
}
