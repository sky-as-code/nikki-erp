package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// A rule or reward write needs update on the promotion as well as its own grant: the right to
// process orders, or to edit an unrelated resource, never implies changing a promotion (BR §9.4).
func TestPromotionChildWritesAssertPromotionUpdate(t *testing.T) {
	for _, tc := range []struct {
		name, grant string
		allowed     bool
	}{
		{"promotion update in org", "update:sales_promotion:org", true},
		{"promotion update named org", "update:sales_promotion:org/org-a", true},
		{"other org", "update:sales_promotion:org/org-b", false},
		{"order rights only", "update:sales_order:org", false},
		{"promotion read only", "read:sales_promotion:org", false},
		{"rule update only", "update:sales_promotion_rule:org", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			org := model.Id("org-a")
			ctx := corectx.NewRequestContext(context.Background())
			orgs := ds.NewSet[model.Id]()
			orgs.Add(org)
			grants := ds.NewSet[string]()
			grants.Add(tc.grant)
			ctx.SetPermissions(corectx.ContextPermissions{
				Principal:    corectx.Principal{Kind: corectx.PrincipalKindService, Id: "svc", OrgId: &org},
				Entitlements: grants, UserOrgIds: orgs,
			})
			cErrs := assertPromotionUpdate(ctx, &org)
			if tc.allowed {
				require.Nil(t, cErrs)
			} else {
				require.NotNil(t, cErrs)
			}
		})
	}
}

// The lifecycle and voucher actions assert seeded codes, never an empty one.
func TestPromotionActionCodes(t *testing.T) {
	require.Equal(t, "activate", PermissionActivate)
	require.Equal(t, "suspend", PermissionSuspend)
	require.Equal(t, "create", PermissionDuplicatePromotion)
	require.Equal(t, "update", PermissionManagePromotion)
	require.Equal(t, "generate", PermissionGenerateVouchers)
	require.Equal(t, "revoke", PermissionRevokeVoucher)
}
