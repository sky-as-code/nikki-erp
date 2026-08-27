// Package app holds the Sales application services: the layer that authorizes a request and then
// delegates the business rule to a domain service.
//
// Authorization happens here and nowhere else. A domain service never checks a permission, and
// never imports this package — so a rule stays callable from a CQRS handler, another module's port
// or a test without dragging the request's identity along with it.
package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
)

// assertPermission answers nil when the caller holds the entitlement, and a client error when it
// does not. The resource code is the schema name, byte-identical to what the engine asserts and to
// what 1007002_sales_iam.sql seeds.
func assertPermission(
	ctx corectx.Context, actionCode string, resourceCode string, scope c.ResourceScope,
) *ft.ClientErrors {
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        scope,
	})
}
