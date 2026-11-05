package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
)

// Resource and action codes for the permission assertions in this package.
//
// The resource code is an alias of the schema name rather than a string literal: the dynamic
// resource engine asserts permissions using the schema name, and iam_resources.code must match it
// byte for byte. A drifted code denies every request with nothing in the response pointing at the
// seed, so the two are tied together here instead of being written out twice.
const (
	settingsRecordResource = models.SettingsRecordSchemaName

	actionRead   = "read"
	actionUpdate = "update"
)

// assertPermission is the module's single authorization call. Authorization belongs to the
// application tier: the domain services below it never assert, so that a caller cannot reach a
// domain service and bypass the check by going around this package.
func assertPermission(ctx corectx.Context, actionCode string, resourceCode string) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        reguard.ResourceScopeTenant,
	})
}

// hasPermission is assertPermission as a yes/no answer, for a read that narrows what it reports
// instead of refusing.
func hasPermission(ctx corectx.Context, actionCode string, resourceCode string) bool {
	return reguard.HasPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        reguard.ResourceScopeTenant,
	})
}

// assertOrgSettingsAccess authorizes an org-level read or write. A request that names no org keeps
// the tenant-scope check and lets the domain service infer the org, as before org_id existed.
//
// A named org is held to the rules of every org-scoped resource: the caller must belong to it, and
// the permission is checked in it so an org-scoped grant answers. It then becomes the acting org,
// which is what the domain service resolves the owner from.
func assertOrgSettingsAccess(ctx corectx.Context, actionCode string, orgId *model.Id) *ft.ClientErrors {
	if orgId == nil {
		return assertPermission(ctx, actionCode, settingsRecordResource)
	}
	if cErr := dynamicresource.AssertCallerInOrg(ctx, *orgId); cErr != nil {
		return cErr
	}
	perm := reguard.PermFor(actionCode, settingsRecordResource, reguard.ResourceScopeOrg).InOrg(orgId)
	if cErr := reguard.AssertPermission(ctx, perm); cErr != nil {
		return cErr
	}
	dyn.WithActingOrg(ctx, *orgId)
	return nil
}
