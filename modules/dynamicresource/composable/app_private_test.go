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
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

const privateCallerId = "user_me"

func newOwnedSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_owned_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50)).PrimaryKey()).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name(basemodel.FieldCreatedBy).DataType(dmodel.FieldDataTypeString(0, 50))).
		Build()
}

func newOwnedAppService(createdBy string, found bool) (*fakeDomainService, CrudApplicationService) {
	tenant := requestguard.ResourceScopeTenant
	domSvc, appSvc := newAppService(newOwnedSchema(), NewAppServiceParam{PermissionScope: &tenant})
	repo := domSvc.repo.(*stubRepository)
	repo.found = found
	repo.record = dmodel.DynamicFields{"id": "r1", basemodel.FieldCreatedBy: createdBy}
	return domSvc, appSvc
}

func privateContext(grants ...string) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	held := ds.NewSet[string]()
	held.AddMany(grants...)
	ctx.SetPermissions(corectx.ContextPermissions{
		Entitlements: held,
		UserId:       model.Id(privateCallerId),
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id(privateCallerId)},
	})
	return ctx
}

func TestPrivateGrantUpdatesTheCallersOwnRecord(t *testing.T) {
	domSvc, appSvc := newOwnedAppService(privateCallerId, true)

	result, err := appSvc.Update(privateContext("update:cmp_owned_resource:private"),
		dmodel.DynamicFields{"id": "r1", "name": "x"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, 1, domSvc.calls)
}

func TestPrivateGrantRefusesSomeoneElsesRecord(t *testing.T) {
	domSvc, appSvc := newOwnedAppService("user_other", true)

	result, err := appSvc.Delete(privateContext("delete:cmp_owned_resource:private"), dmodel.DynamicFields{"id": "r1"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

func TestPrivateFieldGrantNarrowsTheOwnRecordUpdate(t *testing.T) {
	_, appSvc := newOwnedAppService(privateCallerId, true)
	ctx := privateContext("update:cmp_owned_resource.name:private")

	allowed, err := appSvc.Update(ctx, dmodel.DynamicFields{"id": "r1", "name": "x"})
	require.NoError(t, err)
	assert.Zero(t, allowed.ClientErrors.Count())

	refused, err := appSvc.Update(ctx, dmodel.DynamicFields{"id": "r1", basemodel.FieldCreatedBy: "user_other"})
	require.NoError(t, err)
	assert.Positive(t, refused.ClientErrors.Count(), "a private grant on name does not cover created_by")
}

// The record is read only for a caller who holds a private grant at all.
func TestNoPrivateGrantNeverReadsTheRecord(t *testing.T) {
	domSvc, appSvc := newOwnedAppService(privateCallerId, true)

	result, err := appSvc.Update(privateContext("read:cmp_owned_resource:tenant"), dmodel.DynamicFields{"id": "r1"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Nil(t, domSvc.repo.(*stubRepository).fetchedKeys)
}

func TestPrivateReadSearchIsConfinedToTheCallersRecords(t *testing.T) {
	domSvc, appSvc := newOwnedAppService(privateCallerId, true)

	result, err := appSvc.Search(privateContext("read:cmp_owned_resource:private"), dmodel.DynamicFields{})

	require.NoError(t, err)
	require.Zero(t, result.ClientErrors.Count())
	graph, ok := domSvc.seen[fieldNameGraph].(*dmodel.SearchGraph)
	require.True(t, ok, "the search graph carries the owner condition")
	assert.Contains(t, graphFieldNames(graph.ToSearchNode(), nil), basemodel.FieldCreatedBy)
}

func TestTenantReadSearchIsNotConfinedToOwner(t *testing.T) {
	domSvc, appSvc := newOwnedAppService(privateCallerId, true)

	_, err := appSvc.Search(privateContext("read:cmp_owned_resource:tenant", "read:cmp_owned_resource:private"),
		dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Nil(t, domSvc.seen[fieldNameGraph])
}

func TestPrivateScopeNeedsACreatedByColumn(t *testing.T) {
	domSvc, appSvc := newFieldAppService()

	result, err := appSvc.Search(privateContext("read:cmp_field_resource:private"), dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}
