package app

import (
	"context"
	"errors"
	"testing"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/stretchr/testify/require"
)

type orderAuthRepo struct {
	composable.CrudRepository
	result *dyn.OpResult[dmodel.DynamicFields]
	err    error
	t      *testing.T
}

func (r orderAuthRepo) FindByKeys(_ corectx.Context, keys dmodel.DynamicFields) (*dyn.OpResult[dmodel.DynamicFields], error) {
	require.Equal(r.t, dmodel.DynamicFields{"id": "record"}, keys)
	return r.result, r.err
}

func TestOrderPermissionUsesPersistedOrg(t *testing.T) {
	for _, action := range []string{"create", "update", "read"} {
		for _, tc := range []struct {
			name, recordOrg, grant string
			allowed                bool
		}{
			{"own org", "org-a", action + ":sales_order:org", true},
			{"other org", "org-b", action + ":sales_order:org", false},
			{"explicit org", "org-a", action + ":sales_order:org/org-a", true},
			{"wrong explicit org", "org-b", action + ":sales_order:org/org-a", false},
			{"missing grant", "org-a", "read:sales_bill:org", false},
			{"missing record org", "", action + ":sales_order:org", false},
		} {
			t.Run(action+"/"+tc.name, func(t *testing.T) {
				ctx := corectx.NewRequestContext(context.Background())
				org := model.Id("org-a")
				orgs := ds.NewSet[model.Id]()
				orgs.Add(org)
				grants := ds.NewSet[string]()
				grants.Add(tc.grant)
				ctx.SetPermissions(corectx.ContextPermissions{
					Principal:    corectx.Principal{Kind: corectx.PrincipalKindService, Id: "kiosk", OrgId: &org},
					Entitlements: grants, UserOrgIds: orgs,
				})
				repo := orderAuthRepo{t: t, result: &dyn.OpResult[dmodel.DynamicFields]{HasData: true, Data: dmodel.DynamicFields{"org_id": tc.recordOrg}}}
				cErrs, err := assertStoredOrderPermission(ctx, action, "record", repo)
				require.NoError(t, err)
				if tc.allowed {
					require.Nil(t, cErrs)
				} else {
					require.NotNil(t, cErrs)
				}
			})
		}
	}
}

func TestOrderPermissionLookupFailures(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	want := errors.New("read failed")
	_, err := assertStoredOrderPermission(ctx, "create", "record", orderAuthRepo{t: t, err: want})
	require.ErrorIs(t, err, want)
	_, err = assertStoredOrderPermission(ctx, "create", "record", orderAuthRepo{t: t})
	require.Error(t, err)
	cErrs, err := assertStoredOrderPermission(ctx, "create", "record", orderAuthRepo{t: t, result: &dyn.OpResult[dmodel.DynamicFields]{}})
	require.NoError(t, err)
	require.NotNil(t, cErrs)
}

func TestOrderPermissionUnauthenticatedBeforeLookup(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	cErrs, err := assertOrderRecordPermission(ctx, "create", "unregistered", "record")
	require.NoError(t, err)
	require.NotNil(t, cErrs)
}
