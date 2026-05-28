package requestguard

import (
	"fmt"

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
}

type PermissionContext struct {
	IsOwner      bool
	Entitlements []string
}

func AssertPermission(ctx corectx.Context, requiredPerm Perm) *ft.ClientErrors {
	userPerm := ctx.GetPermissions()
	if userPerm.IsOwner {
		return nil
	}

	exactEntitlement := exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, requiredPerm.Scope, requiredPerm.OrgUnitId)
	if userPerm.Entitlements.Contains(exactEntitlement) {
		return nil
	} else if grantedDomain(userPerm, requiredPerm) {
		return nil
	} else if requiredPerm.Scope == ResourceScopeOrg && (grantedExactOrg(userPerm, requiredPerm) ||
		grantedBelongingOrg(userPerm, requiredPerm)) {
		return nil
	}

	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewInsufficientPermissionsError([]string{exactEntitlement}))
	return cErrs
}

// User has domain-level permissions
func grantedDomain(userPerm corectx.ContextPermissions, requiredPerm Perm) bool {
	return userPerm.Entitlements.Contains(omnipotentExpr()) ||
		userPerm.Entitlements.Contains(allActAllRsrcExpr(ResourceScopeDomain, nil)) ||
		userPerm.Entitlements.Contains(thisActAllRsrcExpr(requiredPerm.ActionCode, ResourceScopeDomain, nil)) ||
		userPerm.Entitlements.Contains(allActThisRsrcExpr(requiredPerm.ResourceCode, ResourceScopeDomain, nil)) ||
		userPerm.Entitlements.Contains(exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, ResourceScopeDomain, nil))
}

// // User has permissions in this exact Org Unit
// func grantedExactOrgUnit(userPerm corectx.ContextPermissions, requiredPerm Perm) bool {
// 	return userPerm.Entitlements.Contains(allActAllRsrcExpr("orgunit", requiredPerm.OrgUnitId)) ||
// 		userPerm.Entitlements.Contains(thisActAllRsrcExpr(requiredPerm.ActionCode, "orgunit", requiredPerm.OrgUnitId)) ||
// 		userPerm.Entitlements.Contains(allActThisRsrcExpr(requiredPerm.ResourceCode, "orgunit", requiredPerm.OrgUnitId)) ||
// 		userPerm.Entitlements.Contains(exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, "orgunit", requiredPerm.OrgUnitId))
// }

// // User has permissions in all Org Units they belong to,
// // and user belongs to the same Org Unit as the resource
// func grantedBelongingOrgUnit(userPerm corectx.ContextPermissions, requiredPerm Perm) bool {
// 	return userPerm.OrgUnitId != nil && requiredPerm.OrgUnitId != nil && *userPerm.OrgUnitId == *requiredPerm.OrgUnitId &&
// 		(userPerm.Entitlements.Contains(allActAllRsrcExpr("orgunit", nil)) ||
// 			userPerm.Entitlements.Contains(thisActAllRsrcExpr(requiredPerm.ActionCode, "orgunit", nil)) ||
// 			userPerm.Entitlements.Contains(allActThisRsrcExpr(requiredPerm.ResourceCode, "orgunit", nil)) ||
// 			userPerm.Entitlements.Contains(exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, "orgunit", nil)))
// }

// User has permissions in all Organizations they belong to,
// and user belongs to the same Org as the resource
func grantedBelongingOrg(userPerm corectx.ContextPermissions, requiredPerm Perm) bool {
	return requiredPerm.OrgId != nil && userPerm.UserOrgIds.Contains(*requiredPerm.OrgId) &&
		(userPerm.Entitlements.Contains(allActAllRsrcExpr(ResourceScopeOrg, nil)) ||
			userPerm.Entitlements.Contains(thisActAllRsrcExpr(requiredPerm.ActionCode, ResourceScopeOrg, nil)) ||
			userPerm.Entitlements.Contains(allActThisRsrcExpr(requiredPerm.ResourceCode, ResourceScopeOrg, nil)) ||
			userPerm.Entitlements.Contains(exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, ResourceScopeOrg, nil)))
}

func grantedExactOrg(userPerm corectx.ContextPermissions, requiredPerm Perm) bool {
	return requiredPerm.OrgId != nil &&
		(userPerm.Entitlements.Contains(allActAllRsrcExpr(ResourceScopeOrg, requiredPerm.OrgId)) ||
			userPerm.Entitlements.Contains(thisActAllRsrcExpr(requiredPerm.ActionCode, ResourceScopeOrg, requiredPerm.OrgId)) ||
			userPerm.Entitlements.Contains(allActThisRsrcExpr(requiredPerm.ResourceCode, ResourceScopeOrg, requiredPerm.OrgId)) ||
			userPerm.Entitlements.Contains(exactExpr(requiredPerm.ActionCode, requiredPerm.ResourceCode, ResourceScopeOrg, requiredPerm.OrgId)))
}

func exactExpr(actionCode string, resourceCode string, scope ResourceScope, scopeId *model.Id) string {
	if scopeId != nil {
		return fmt.Sprintf("%s:%s:%s/%s", actionCode, resourceCode, string(scope), *scopeId)
	}
	return fmt.Sprintf("%s:%s:%s", actionCode, resourceCode, scope)
}

func allActThisRsrcExpr(resourceCode string, scope ResourceScope, scopeId *model.Id) string {
	if scopeId != nil {
		return fmt.Sprintf("*:%s:%s/%s", resourceCode, string(scope), *scopeId)
	}
	return fmt.Sprintf("*:%s:%s", resourceCode, string(scope))
}

func thisActAllRsrcExpr(actionCode string, scope ResourceScope, scopeId *model.Id) string {
	if scopeId != nil {
		return fmt.Sprintf("%s:*:%s/%s", actionCode, string(scope), *scopeId)
	}
	return fmt.Sprintf("%s:*:%s", actionCode, string(scope))
}

func allActAllRsrcExpr(scope ResourceScope, scopeId *model.Id) string {
	if scopeId != nil {
		return fmt.Sprintf("*:*:%s/%s", string(scope), *scopeId)
	}
	return fmt.Sprintf("*:*:%s", string(scope))
}

func omnipotentExpr() string {
	return "*:*:*"
}
