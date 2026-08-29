// Package app holds the Accounting application services: the layer that authorizes a request and
// then delegates to a domain service.
//
// Authorization happens here and nowhere else; a domain service never checks a permission and never
// imports this package, so a rule stays callable without a request identity.
package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	c "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// assertPermission answers nil when the caller holds the entitlement. The resource code is the
// schema name, byte-identical to what the engine asserts and to what 1003002_accounting_iam.sql
// seeds.
func assertPermission(
	ctx corectx.Context, actionCode string, resourceCode string, scope c.ResourceScope,
) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        scope,
	})
}
