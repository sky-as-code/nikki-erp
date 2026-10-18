package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

const (
	testResourceId = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY1")
	testReadId     = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY2")
	testForeignId  = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY3")
	testRoleOrgId  = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY4")
	testGrantField = "entitlements.add[0]"
)

func testResource(maxScope string) domain.Resource {
	return *domain.NewResourceFrom(dmodel.DynamicFields{
		domain.ResourceFieldCode:     "iam_user",
		domain.ResourceFieldMaxScope: maxScope,
		domain.ResourceEdgeActions: []dmodel.DynamicFields{
			{domain.ActionFieldId: testReadId, domain.ActionFieldCode: "read"},
		},
	})
}

func TestPlanGrantOnResource_TenantResourceWithAction(t *testing.T) {
	actionId := testReadId
	grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId, ActionId: &actionId}

	plan, cErr := planGrantOnResource(testResource("tenant"), nil, grant, testGrantField)

	require.Nil(t, cErr)
	assert.Equal(t, "read:iam_user:tenant", plan.expression)
	assert.Nil(t, plan.orgId)
}

func TestPlanGrantOnResource_NilActionIsWildcard(t *testing.T) {
	grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId}

	plan, cErr := planGrantOnResource(testResource("tenant"), nil, grant, testGrantField)

	require.Nil(t, cErr)
	assert.Equal(t, "*:iam_user:tenant", plan.expression)
}

func TestPlanGrantOnResource_OrgScopeTakesRoleOrg(t *testing.T) {
	orgId := testRoleOrgId
	grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId}

	plan, cErr := planGrantOnResource(testResource("org"), &orgId, grant, testGrantField)

	require.Nil(t, cErr)
	assert.Equal(t, "*:iam_user:org/"+string(testRoleOrgId), plan.expression)
	require.NotNil(t, plan.orgId)
	assert.Equal(t, testRoleOrgId, *plan.orgId)
	assert.Equal(t, testRoleOrgId, *domain.NewEntitlementFrom(plan.command("role")).GetOrgId())
}

func TestPlanGrantOnResource_OrgScopeWithoutRoleOrgIsRefused(t *testing.T) {
	grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId}

	_, cErr := planGrantOnResource(testResource("org"), nil, grant, testGrantField)

	require.NotNil(t, cErr)
	assert.Equal(t, "iam:err_role_org_required", cErr.Key)
	assert.Equal(t, testGrantField, cErr.Field)
}

func TestPlanGrantOnResource_NarrowScopesAreNotGrantable(t *testing.T) {
	orgId := testRoleOrgId
	for _, scope := range []string{"orgunit", "private"} {
		grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId}

		_, cErr := planGrantOnResource(testResource(scope), &orgId, grant, testGrantField)

		require.NotNil(t, cErr, scope)
		assert.Equal(t, "iam:err_scope_not_grantable", cErr.Key, scope)
	}
}

func TestPlanGrantOnResource_ForeignActionIsNotFound(t *testing.T) {
	actionId := testForeignId
	grant := itRole.RoleEntitlementGrant{ResourceId: testResourceId, ActionId: &actionId}

	_, cErr := planGrantOnResource(testResource("tenant"), nil, grant, testGrantField)

	require.NotNil(t, cErr)
	assert.Equal(t, testGrantField+".action_id", cErr.Field)
}

func TestValidateDeltaSize_RefusesOversizedLists(t *testing.T) {
	delta := itRole.RoleEntitlementsDelta{
		Remove: make([]model.Id, itRole.RoleEntitlementsDeltaMaxItems+1),
	}

	cErrs := validateDeltaSize(delta)

	require.Equal(t, 1, cErrs.Count())
	assert.Equal(t, "entitlements.remove", cErrs[0].Field)
}

func TestPrefixFields_NestsUnderTheDeltaItem(t *testing.T) {
	cErrs := ft.ClientErrors{
		*ft.NewNotFoundError("resource_id"),
		*ft.NewNotFoundError(""),
	}

	out := prefixFields(cErrs, testGrantField)

	assert.Equal(t, testGrantField+".resource_id", out[0].Field)
	assert.Equal(t, testGrantField, out[1].Field)
}

func TestValidateDeltaSize_RefusesOversizedUpdates(t *testing.T) {
	delta := itRole.RoleEntitlementsDelta{
		Update: make([]itRole.RoleEntitlementUpdate, itRole.RoleEntitlementsDeltaMaxItems+1),
	}

	cErrs := validateDeltaSize(delta)

	require.Equal(t, 1, cErrs.Count())
	assert.Equal(t, "entitlements.update", cErrs[0].Field)
}

func TestValidateEntitlementUpdates_RefusesMissingRemovedAndRepeatedIds(t *testing.T) {
	delta := itRole.RoleEntitlementsDelta{
		Remove: []model.Id{testForeignId},
		Update: []itRole.RoleEntitlementUpdate{
			{Id: testReadId},
			{Id: ""},
			{Id: testForeignId},
			{Id: testReadId},
		},
	}

	cErrs := validateEntitlementUpdates(delta)

	require.Equal(t, 3, cErrs.Count())
	assert.Equal(t, "entitlements.update[1].id", cErrs[0].Field)
	assert.Equal(t, "iam:err_entitlement_update_removed", cErrs[1].Key)
	assert.Equal(t, "entitlements.update[2]", cErrs[1].Field)
	assert.Equal(t, "iam:err_entitlement_update_repeated", cErrs[2].Key)
	assert.Equal(t, "entitlements.update[3]", cErrs[2].Field)
}

func TestAssertUpdatesOwned_RefusesForeignEntitlements(t *testing.T) {
	updates := []itRole.RoleEntitlementUpdate{{Id: testReadId}, {Id: testForeignId}}
	etags := map[model.Id]model.Etag{testReadId: "etag-1"}

	cErrs := assertUpdatesOwned(updates, etags)

	require.Equal(t, 1, cErrs.Count())
	assert.Equal(t, "iam:err_entitlement_not_in_role", cErrs[0].Key)
	assert.Equal(t, "entitlements.update[1]", cErrs[0].Field)
}

func TestEntitlementUpdateCommand_CarriesOnlyIdEtagAndDescription(t *testing.T) {
	description := "Read users"

	cmd := entitlementUpdateCommand(itRole.RoleEntitlementUpdate{Id: testReadId, Description: &description}, "etag-1")

	fields := cmd
	assert.Equal(t, testReadId, *domain.NewEntitlementFrom(cmd).GetId())
	assert.Equal(t, "etag-1", fields["etag"])
	assert.Equal(t, description, fields[domain.EntitlementFieldDescription])
	assert.Len(t, fields, 3)
}
