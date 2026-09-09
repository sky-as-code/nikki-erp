package requestguard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// A user and a service authorize through the same path, differing only in the identity and the
// entitlements they were granted - never in whether they are checked at all. These tests are the
// executable form of that rule.

var (
	principalOrgA = model.Id("01JQZ0X000000000000000000A")
	principalOrgB = model.Id("01JQZ0X000000000000000000B")
)

func contextFor(principal corectx.Principal, grants []string, orgIds ...model.Id) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	entitlements := ds.NewSet[string]()
	entitlements.AddMany(grants...)
	orgs := ds.NewSet[model.Id]()
	orgs.AddMany(orgIds...)

	ctx.SetPermissions(corectx.ContextPermissions{
		Entitlements: entitlements,
		UserId:       principal.Id,
		Principal:    principal,
		UserOrgIds:   orgs,
	})
	return ctx
}

func userPrincipal() corectx.Principal {
	return corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("01JQZ0X0000000000000000001")}
}

func servicePrincipal(orgId *model.Id) corectx.Principal {
	return corectx.Principal{
		Kind:        corectx.PrincipalKindService,
		Id:          model.Id("01JQZ0X0000000000000000002"),
		OrgId:       orgId,
		DisplayName: "svc_billing",
	}
}

func hasKey(cErrs *ft.ClientErrors, key string) bool {
	if cErrs == nil {
		return false
	}
	for _, item := range *cErrs {
		if item.Key == key {
			return true
		}
	}
	return false
}

/* User principal */

func TestUserWithTheEntitlementIsAllowed(t *testing.T) {
	ctx := contextFor(userPrincipal(), []string{"read:iam_user:tenant"})

	cErrs := AssertPermission(ctx, PermFor("read", "iam_user", ResourceScopeTenant))

	assert.Nil(t, cErrs)
}

func TestUserWithoutTheEntitlementIsForbidden(t *testing.T) {
	ctx := contextFor(userPrincipal(), []string{"read:sales_order:tenant"})

	cErrs := AssertPermission(ctx, PermFor("read", "iam_user", ResourceScopeTenant))

	require.NotNil(t, cErrs)
	assert.True(t, hasKey(cErrs, ft.ErrorKey("err_insufficient_permissions")),
		"a known caller lacking a grant is forbidden, not unauthenticated")
}

/* Service principal */

func TestServiceWithTheEntitlementIsAllowed(t *testing.T) {
	ctx := contextFor(servicePrincipal(&principalOrgA), []string{"update:sales_order:tenant"})

	cErrs := AssertPermission(ctx, PermFor("update", "sales_order", ResourceScopeTenant))

	assert.Nil(t, cErrs)
}

func TestServiceWithoutTheEntitlementIsForbidden(t *testing.T) {
	ctx := contextFor(servicePrincipal(&principalOrgA), nil)

	cErrs := AssertPermission(ctx, PermFor("update", "sales_order", ResourceScopeTenant))

	require.NotNil(t, cErrs)
	assert.True(t, hasKey(cErrs, ft.ErrorKey("err_insufficient_permissions")))
}

// The rule the CR is most concerned with: being internal is not a permission. A service with an
// empty entitlement set is refused exactly as a user would be.
func TestServiceHasNoImplicitSuperuserPrivilege(t *testing.T) {
	service := contextFor(servicePrincipal(&principalOrgA), nil)
	user := contextFor(userPrincipal(), nil)

	required := PermFor("delete", "sales_order", ResourceScopeTenant)

	assert.NotNil(t, AssertPermission(service, required))
	assert.NotNil(t, AssertPermission(user, required),
		"the two must be refused on the same grounds; neither kind is privileged")
}

// A grant naming one org must not answer for another, whoever holds it.
func TestServiceGrantForOneOrgDoesNotAnswerForAnother(t *testing.T) {
	grant := BuildExpression("update", "sales_order", ResourceScopeOrg, &principalOrgA)
	ctx := contextFor(servicePrincipal(&principalOrgA), []string{grant}, principalOrgA)

	allowed := AssertPermission(ctx, PermFor("update", "sales_order", ResourceScopeOrg).InOrg(&principalOrgA))
	refused := AssertPermission(ctx, PermFor("update", "sales_order", ResourceScopeOrg).InOrg(&principalOrgB))

	assert.Nil(t, allowed)
	assert.NotNil(t, refused, "org isolation holds for a service principal too")
}

/* Missing principal */

// Nothing authenticated the caller, so the answer is "unauthenticated" rather than "you lack a
// grant" - the distinction that stops an administrator hunting for an entitlement to add when the
// real fault is a job or consumer that never established a principal.
func TestAssertPermissionFailsClosedWithoutAPrincipal(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	cErrs := AssertPermission(ctx, PermFor("read", "iam_user", ResourceScopeTenant))

	require.NotNil(t, cErrs)
	assert.True(t, hasKey(cErrs, ft.ErrorKey("err_unauthenticated")),
		"a missing principal must report unauthenticated, not insufficient permissions")
}

// Ownership does not rescue an execution that nobody authenticated: the IsOwner short-circuit sits
// behind the principal check, so a zero context cannot claim it.
func TestOwnerShortCircuitDoesNotBypassThePrincipalCheck(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{IsOwner: true})

	cErrs := AssertPermission(ctx, PermFor("delete", "iam_user", ResourceScopeTenant))

	require.NotNil(t, cErrs)
	assert.True(t, hasKey(cErrs, ft.ErrorKey("err_unauthenticated")))
}

// An authenticated owner is still an owner.
func TestOwnerWithAPrincipalIsAllowed(t *testing.T) {
	ctx := contextFor(userPrincipal(), nil)
	perms := ctx.GetPermissions()
	perms.IsOwner = true
	ctx.SetPermissions(perms)

	cErrs := AssertPermission(ctx, PermFor("delete", "iam_user", ResourceScopeTenant))

	assert.Nil(t, cErrs)
}
