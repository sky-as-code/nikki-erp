// Package app holds the Sales application services. Authorization happens here and nowhere else:
// domain services never check permissions nor import this package, so a rule stays callable from a
// CQRS handler, another module's port or a test without carrying the request identity.
package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
)

// assertPermission checks the caller's entitlement. The resource code is the schema name,
// byte-identical to what the engine asserts and to what 1007002_sales_iam.sql seeds.
func assertPermission(
	ctx corectx.Context, actionCode string, resourceCode string, scope c.ResourceScope,
) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        scope,
	})
}
