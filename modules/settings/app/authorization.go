package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
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
		Scope:        reguard.ResourceScopeDomain,
	})
}
