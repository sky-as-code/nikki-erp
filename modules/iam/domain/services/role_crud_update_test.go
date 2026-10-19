package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

const testRoleId = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY6")

// fakeEntitlementRepo answers every search with the role's stored entitlements.
type fakeEntitlementRepo struct {
	itEnt.EntitlementCrudRepository
	held []dmodel.DynamicFields
}

func (this *fakeEntitlementRepo) Search(corectx.Context, dyn.RepoSearchParam) (*composable.SearchResult, error) {
	return &composable.SearchResult{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.held, Total: len(this.held)},
		HasData: len(this.held) > 0,
	}, nil
}

type fakeResourceRepo struct {
	itRes.ResourceCrudRepository
	resource dmodel.DynamicFields
}

func (this *fakeResourceRepo) GetOne(corectx.Context, dyn.RepoGetOneParam) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return &dyn.OpResult[dmodel.DynamicFields]{Data: this.resource, HasData: this.resource != nil}, nil
}

type fakeEntitlementSvc struct {
	itEnt.EntitlementCrudDomainService
	created []itEnt.CreateEntitlementCommand
}

func (this *fakeEntitlementSvc) CreateDeferRebuild(
	_ corectx.Context, cmd itEnt.CreateEntitlementCommand,
) (*itEnt.CreateEntitlementResult, error) {
	this.created = append(this.created, cmd)
	return &itEnt.CreateEntitlementResult{Data: cmd, HasData: true}, nil
}

type roleUpdateFixture struct {
	svc      *RoleCrudDomainServiceImpl
	base     *fakeCrudDomain
	ents     *fakeEntitlementRepo
	entSvc   *fakeEntitlementSvc
	permRepo *fakePermRepo
}

// newRoleUpdateFixture stores a role without an org and a tenant-scoped resource with a read action.
func newRoleUpdateFixture() *roleUpdateFixture {
	fx := &roleUpdateFixture{
		base:     newFakeCrudDomain(),
		ents:     &fakeEntitlementRepo{},
		entSvc:   &fakeEntitlementSvc{},
		permRepo: &fakePermRepo{},
	}
	fx.base.repo.rows = map[string]dmodel.DynamicFields{string(testRoleId): {"id": testRoleId}}
	resources := &fakeResourceRepo{resource: testResource("tenant").GetFieldData()}
	fx.svc = NewRoleCrudDomainService(fx.base, RoleCrudDeps{
		EntitlementRepo: fx.ents,
		EntitlementSvc:  fx.entSvc,
		ResourceRepo:    resources,
		PermRepo:        fx.permRepo,
	}).(*RoleCrudDomainServiceImpl)
	return fx
}

func roleUpdateCommand(delta map[string]any) composable.UpdateCommand {
	return dmodel.DynamicFields{
		"id":                               string(testRoleId),
		"etag":                             "etag-1",
		domain.RoleFieldName:               "Renamed",
		itRole.UpdateRoleFieldEntitlements: delta,
	}
}

func readGrant() map[string]any {
	return map[string]any{"add": []any{map[string]any{"resource_id": testResourceId, "action_id": testReadId}}}
}

func (this *roleUpdateFixture) assertRolledBack(t *testing.T, result *composable.MutateResult, key string) {
	t.Helper()
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, key, result.ClientErrors[0].Key)
	assert.Len(t, this.base.repo.updates, 1, "the role row is written before the delta is checked")
	assert.False(t, this.base.repo.tranx.committed, "a refused delta must not commit the role row")
	assert.True(t, this.base.repo.tranx.rolledBack)
	assert.Empty(t, this.entSvc.created)
	assert.Empty(t, this.permRepo.rebuiltRoles)
}

func TestRoleUpdate_AlreadyGrantedIsRefusedAndRollsBack(t *testing.T) {
	fx := newRoleUpdateFixture()
	fx.ents.held = []dmodel.DynamicFields{{"id": testForeignId, domain.EntitlementFieldExpression: "read:iam_user:tenant"}}

	result, err := fx.svc.Update(testContext(), roleUpdateCommand(readGrant()))

	require.NoError(t, err)
	fx.assertRolledBack(t, result, "iam:err_entitlement_already_granted")
}

func TestRoleUpdate_OrgScopedGrantOnRoleWithoutOrgIsRefusedAndRollsBack(t *testing.T) {
	fx := newRoleUpdateFixture()
	fx.svc.resourceRepo = &fakeResourceRepo{resource: testResource("org").GetFieldData()}

	result, err := fx.svc.Update(testContext(), roleUpdateCommand(readGrant()))

	require.NoError(t, err)
	fx.assertRolledBack(t, result, "iam:err_role_org_required")
}

func TestRoleUpdate_RemovingAnotherRolesEntitlementIsRefusedAndRollsBack(t *testing.T) {
	fx := newRoleUpdateFixture()

	result, err := fx.svc.Update(testContext(), roleUpdateCommand(map[string]any{"remove": []any{testForeignId}}))

	require.NoError(t, err)
	fx.assertRolledBack(t, result, "iam:err_entitlement_not_in_role")
}

func TestRoleUpdate_AcceptedDeltaCommitsWithOneRebuild(t *testing.T) {
	fx := newRoleUpdateFixture()

	result, err := fx.svc.Update(testContext(), roleUpdateCommand(readGrant()))

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Len(t, fx.base.repo.updates, 1)
	assert.Len(t, fx.entSvc.created, 1)
	assert.Equal(t, []model.Id{testRoleId}, fx.permRepo.rebuiltRoles)
	assert.True(t, fx.base.repo.tranx.committed)
}
