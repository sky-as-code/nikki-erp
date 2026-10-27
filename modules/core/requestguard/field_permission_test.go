package requestguard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func contextHolding(grants ...string) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	held := ds.NewSet[string]()
	held.AddMany(grants...)
	ctx.SetPermissions(corectx.ContextPermissions{
		Entitlements: held,
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("01JQZ0X0000000000000000001")},
	})
	return ctx
}

func missingEntitlements(t *testing.T, cErrs *ft.ClientErrors) []string {
	require.NotNil(t, cErrs)
	require.Equal(t, 1, cErrs.Count())
	missing, ok := (*cErrs)[0].Vars["entitlements"].([]string)
	require.True(t, ok)
	return missing
}

var updateUser = PermFor("update", "iam_user", ResourceScopeTenant)

func TestAssertFieldsPermission_ResourceGrantCoversEveryField(t *testing.T) {
	ctx := contextHolding("update:iam_user:tenant")
	fields := dmodel.DynamicFields{"display_name": "x", "email": "y"}
	assert.Nil(t, AssertFieldsPermission(ctx, updateUser, fields))
}

func TestAssertFieldsPermission_FieldGrantCoversOnlyItsField(t *testing.T) {
	ctx := contextHolding("update:iam_user.display_name:tenant")

	assert.Nil(t, AssertFieldsPermission(ctx, updateUser, dmodel.DynamicFields{"display_name": "x"}))

	refused := AssertFieldsPermission(ctx, updateUser, dmodel.DynamicFields{
		"display_name": "x", "email": "y", "status": "z",
	})
	assert.Equal(t, []string{"update:iam_user.email:tenant", "update:iam_user.status:tenant"},
		missingEntitlements(t, refused), "every missing field is reported, sorted")
}

func TestAssertFieldsPermission_NothingToAuthorize(t *testing.T) {
	assert.Nil(t, AssertFieldsPermission(contextHolding(), updateUser, dmodel.DynamicFields{}))
	assert.NotNil(t, AssertFieldsPermission(corectx.NewRequestContext(context.Background()), updateUser,
		dmodel.DynamicFields{}), "an unauthenticated caller is refused even with nothing to write")
}

func TestReadableFields(t *testing.T) {
	names := []string{"id", "display_name", "email"}
	read := PermFor("read", "iam_user", ResourceScopeTenant)

	assert.Equal(t, names, ReadableFields(contextHolding("read:iam_user:tenant"), read, names))
	assert.Equal(t, []string{"display_name"},
		ReadableFields(contextHolding("read:iam_user.display_name:tenant"), read, names))
	assert.Empty(t, ReadableFields(contextHolding("update:iam_user.email:tenant"), read, names))
}

func TestAssertAnyFieldPermission(t *testing.T) {
	names := []string{"display_name", "email"}

	assert.Nil(t, AssertAnyFieldPermission(contextHolding("update:iam_user:tenant"), updateUser, names))
	assert.Nil(t, AssertAnyFieldPermission(contextHolding("update:iam_user.email:tenant"), updateUser, names))

	refused := AssertAnyFieldPermission(contextHolding("update:iam_user.status:tenant"), updateUser, names)
	assert.Equal(t, []string{"update:iam_user:tenant"}, missingEntitlements(t, refused),
		"a refusal names the resource-level grant")
}

func TestFieldPermission_OrgScoped(t *testing.T) {
	orgId := model.Id("ORG1")
	read := PermFor("read", "iam_user", ResourceScopeOrg).InOrg(&orgId)

	ctx := contextHolding("read:iam_user.email:org/ORG1")
	assert.Equal(t, []string{"email"}, ReadableFields(ctx, read, []string{"display_name", "email"}))

	other := model.Id("ORG2")
	assert.Empty(t, ReadableFields(ctx, read.InOrg(&other), []string{"display_name", "email"}))
}
