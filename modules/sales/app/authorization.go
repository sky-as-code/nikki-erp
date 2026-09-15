// Package app holds the Sales application services. Authorization happens here and nowhere else:
// domain services never check permissions nor import this package, so a rule stays callable from a
// CQRS handler, another module's port or a test without carrying the request identity.
package app

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
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

// The in-process order API authorizes against persisted ownership, not the caller's
// selected org. Create uses the sales point; later actions use the order/fulfillment.
func assertOrderRecordPermission(ctx corectx.Context, action, schema, id string) (*ft.ClientErrors, error) {
	if ctx.GetPermissions().Principal.IsZero() {
		return assertPermission(ctx, action, c.SalesOrderResource, c.ResourceScopeOrg), nil
	}
	repo, err := services.RepositoryFor(schema)
	if err != nil {
		return nil, err
	}
	return assertStoredOrderPermission(ctx, action, id, repo)
}

func assertStoredOrderPermission(ctx corectx.Context, action, id string, repo composable.CrudRepository) (*ft.ClientErrors, error) {
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{"id": id})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("Sales authorization lookup returned no result")
	}
	if found.ClientErrors.Count() > 0 {
		return &found.ClientErrors, nil
	}
	if !found.HasData {
		return &ft.ClientErrors{*ft.NewNotFoundError("id")}, nil
	}
	orgId := found.Data.GetModelId("org_id")
	if orgId == nil || *orgId == "" {
		return &ft.ClientErrors{*ft.NewInsufficientPermissionsError([]string{
			reguard.BuildExpression(action, c.SalesOrderResource, c.ResourceScopeOrg, nil),
		})}, nil
	}
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode: action, ResourceCode: c.SalesOrderResource,
		Scope: c.ResourceScopeOrg, OrgId: orgId,
	}), nil
}
