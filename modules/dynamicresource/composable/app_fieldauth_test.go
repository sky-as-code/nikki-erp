package composable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const fieldSchemaName = "cmp_field_resource"

func newFieldSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel(fieldSchemaName).
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50)).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name("secret").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

func newFieldAppService() (*fakeDomainService, CrudApplicationService) {
	tenant := requestguard.ResourceScopeTenant
	return newAppService(newFieldSchema(), NewAppServiceParam{PermissionScope: &tenant})
}

// grantedContext is a non-owner holding exactly the given entitlements.
func grantedContext(grants ...string) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	held := ds.NewSet[string]()
	held.AddMany(grants...)
	ctx.SetPermissions(corectx.ContextPermissions{
		Entitlements: held,
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("field_test")},
	})
	return ctx
}

func TestUpdateWithAFieldGrantWritesOnlyThatField(t *testing.T) {
	domSvc, appSvc := newFieldAppService()
	ctx := grantedContext("update:cmp_field_resource.name:tenant")

	allowed, err := appSvc.Update(ctx, dmodel.DynamicFields{"id": "r1", "name": "x"})
	require.NoError(t, err)
	assert.Zero(t, allowed.ClientErrors.Count(), "the id addresses the record and needs no grant")

	refused, err := appSvc.Update(ctx, dmodel.DynamicFields{"id": "r1", "name": "x", "secret": "y"})
	require.NoError(t, err)
	require.Equal(t, 1, refused.ClientErrors.Count())
	assert.Equal(t, []string{"update:cmp_field_resource.secret:tenant"}, refused.ClientErrors[0].Vars["entitlements"])
	assert.Equal(t, 1, domSvc.calls, "the refused update never reaches the domain service")
}

func TestCreateWithAResourceGrantWritesEveryField(t *testing.T) {
	_, appSvc := newFieldAppService()
	result, err := appSvc.Create(grantedContext("create:cmp_field_resource:tenant"),
		dmodel.DynamicFields{"name": "x", "secret": "y"})
	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
}

func TestDeleteIgnoresFieldGrants(t *testing.T) {
	domSvc, appSvc := newFieldAppService()
	result, err := appSvc.Delete(grantedContext("delete:cmp_field_resource.name:tenant"),
		dmodel.DynamicFields{"id": "r1"})
	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestSearchRefusesAFilterOnAnUnreadableField(t *testing.T) {
	domSvc, appSvc := newFieldAppService()
	ctx := grantedContext("read:cmp_field_resource.name:tenant")

	graph := dmodel.NewSearchGraph().NewCondition("secret", dmodel.Equals, "x")
	refused, err := appSvc.Search(ctx, dmodel.DynamicFields{fieldNameGraph: graph})
	require.NoError(t, err)
	assert.Positive(t, refused.ClientErrors.Count())

	sorted := dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "x").OrderBy("secret")
	refused, err = appSvc.Search(ctx, dmodel.DynamicFields{fieldNameGraph: sorted})
	require.NoError(t, err)
	assert.Positive(t, refused.ClientErrors.Count(), "sorting discloses the order of the values")

	allowed, err := appSvc.Search(ctx, dmodel.DynamicFields{
		fieldNameGraph: dmodel.NewSearchGraph().NewCondition("name", dmodel.Equals, "x"),
	})
	require.NoError(t, err)
	assert.Zero(t, allowed.ClientErrors.Count())
	assert.Equal(t, 1, domSvc.calls)
}

func TestSearchRefusesACallerWithNoReadGrantAtAll(t *testing.T) {
	_, appSvc := newFieldAppService()
	result, err := appSvc.Search(grantedContext("update:cmp_field_resource.name:tenant"), dmodel.DynamicFields{})
	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
}

func TestNarrowReadFieldsMasksUnreadableFields(t *testing.T) {
	_, appSvc := newFieldAppService()
	record := dmodel.DynamicFields{"id": "r1", "name": "x", "secret": "y"}

	masked := appSvc.NarrowReadFields(grantedContext("read:cmp_field_resource.name:tenant"), nil, nil, record)

	assert.Equal(t, []string{"secret"}, masked)
	assert.Equal(t, dmodel.DynamicFields{"id": "r1", "name": "x", "secret": nil}, record)
}

func TestNarrowReadFieldsLeavesAResourceReaderAlone(t *testing.T) {
	_, appSvc := newFieldAppService()
	record := dmodel.DynamicFields{"id": "r1", "name": "x", "secret": "y"}

	masked := appSvc.NarrowReadFields(grantedContext("read:cmp_field_resource:tenant"), nil, []string{"name"}, record)

	assert.Empty(t, masked)
	assert.Equal(t, "y", record["secret"])
}

func TestNarrowReadFieldsReportsUnreadableDesiredFields(t *testing.T) {
	_, appSvc := newFieldAppService()
	record := dmodel.DynamicFields{"id": "r1", "name": "x"}

	masked := appSvc.NarrowReadFields(grantedContext("read:cmp_field_resource.name:tenant"), nil,
		[]string{"name", "secret"}, record)

	assert.Equal(t, []string{"secret"}, masked)
	assert.NotContains(t, record, "secret", "a field absent from the record is reported, not added")
}
