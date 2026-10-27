package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// principalRows holds one user (no type: the column default) and one service.
func principalRows() map[string]dmodel.DynamicFields {
	return map[string]dmodel.DynamicFields{
		"U1": {"id": "U1"},
		"S1": {"id": "S1", domain.UserFieldType: string(domain.UserTypeService)},
	}
}

// searchCapture is the composable default recording the search it was handed.
type searchCapture struct {
	*fakeCrudDomain
	searched composable.SearchQuery
}

func (this *searchCapture) Search(
	_ corectx.Context, params composable.SearchQuery, _ ...corecrud.ServiceSearchOptions,
) (*composable.SearchResult, error) {
	this.searched = params
	return &composable.SearchResult{HasData: true}, nil
}

func (this *searchCapture) Schema() *dmodel.ModelSchema {
	return dmodel.DefineModel("iam_principal_test").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("display_name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

func newTypedBase() *searchCapture {
	base := &searchCapture{fakeCrudDomain: newFakeCrudDomain()}
	base.repo.rows = principalRows()
	return base
}

func graphJson(t *testing.T, params composable.SearchQuery) string {
	encoded, err := json.Marshal(params["graph"])
	require.NoError(t, err)
	return string(encoded)
}

func TestUserDomainServiceHidesServices(t *testing.T) {
	base := newTypedBase()
	svc := NewUserCrudDomainService(base, nil, &fakePermRepo{}, nil)

	got, err := svc.GetById(testContext(), dmodel.DynamicFields{"id": "S1"})
	require.NoError(t, err)
	assert.False(t, got.HasData, "a service id is not found on iam_user")

	updated, err := svc.Update(testContext(), dmodel.DynamicFields{"id": "S1", "display_name": "x"})
	require.NoError(t, err)
	require.Equal(t, 1, updated.ClientErrors.Count())
	assert.Empty(t, base.repo.updates)

	_, err = svc.Search(testContext(), composable.SearchQuery{})
	require.NoError(t, err)
	assert.Contains(t, graphJson(t, base.searched), `"type","=","user"`)
}

func TestServiceDomainServiceHidesUsers(t *testing.T) {
	base := newTypedBase()
	svc := NewServiceCrudDomainService(base, &fakePermRepo{}, nil)

	got, err := svc.GetById(testContext(), dmodel.DynamicFields{"id": "U1"})
	require.NoError(t, err)
	assert.False(t, got.HasData, "a user id is not found on iam_service")

	deleted, err := svc.Delete(testContext(), dmodel.DynamicFields{"id": "U1"})
	require.NoError(t, err)
	assert.Equal(t, 1, deleted.ClientErrors.Count())
	assert.Zero(t, base.deleted)

	_, err = svc.Search(testContext(), composable.SearchQuery{
		"graph":  map[string]any{"if": []any{"display_name", "=", "x"}},
		"fields": []string{"display_name", "email"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"display_name"}, base.searched["fields"], "a user-only column is never projected")
	assert.Contains(t, graphJson(t, base.searched), `"type","=","service"`)
	assert.Contains(t, graphJson(t, base.searched), `"display_name"`, "the caller's own filter is kept")
}

func TestServiceCreateWritesTheServiceType(t *testing.T) {
	entity := composable.NewDynamicEntityFrom(dmodel.DynamicFields{"display_name": "bot"})
	result, err := setServiceType(nil)(testContext(), entity)
	require.NoError(t, err)
	assert.Equal(t, string(domain.UserTypeService), result.GetFieldData()[domain.UserFieldType])
}

func TestUserCreateForcesTheUserType(t *testing.T) {
	base := newTypedBase()
	svc := NewUserCrudDomainService(base, &fakeUserSettings{}, &fakePermRepo{}, nil)
	cmd := dmodel.DynamicFields{"display_name": "x", domain.UserFieldType: string(domain.UserTypeService)}

	result, err := svc.Create(testContext(), cmd)
	require.NoError(t, err)
	require.True(t, result.HasData)
	assert.Equal(t, string(domain.UserTypeUser), base.created[domain.UserFieldType])
}

func TestOrgUnitManageUsersRefusesAService(t *testing.T) {
	svc, userRepo := newOrgUnitService(map[string]dmodel.DynamicFields{"U1": {}}, principalRows())

	result, err := svc.ManageUsers(testContext(), dmodel.DynamicFields{
		"id": "U1", composable.FieldM2mAdd: []any{"S1"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, ft.ErrorKey(errPrincipalNotUser, "iam"), result.ClientErrors[0].Key)
	assert.Empty(t, userRepo.updates)
}

func TestRefuseServiceInOwnerField(t *testing.T) {
	userRepo := &stubCrudRepo{rows: principalRows()}
	guard := refuseServiceIn(userRepo, domain.GroupFieldOwnerId)
	vErrs := ft.ClientErrors{}

	human := composable.NewDynamicEntityFrom(dmodel.DynamicFields{domain.GroupFieldOwnerId: "U1"})
	require.NoError(t, guard(testContext(), human, nil, &vErrs))
	assert.Zero(t, vErrs.Count())

	machine := composable.NewDynamicEntityFrom(dmodel.DynamicFields{domain.GroupFieldOwnerId: "S1"})
	require.NoError(t, guard(testContext(), machine, nil, &vErrs))
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, domain.GroupFieldOwnerId, vErrs[0].Field)
}

func TestRoleAssignmentsRefuseTheOtherPrincipalType(t *testing.T) {
	base := newFakeCrudDomain()
	base.repo.rows = map[string]dmodel.DynamicFields{"R1": {"id": "R1"}}
	svc := &RoleCrudDomainServiceImpl{
		CrudDomainService: base, roleRepo: base.repo, permRepo: &fakePermRepo{},
		userRepo: &stubCrudRepo{rows: principalRows()},
	}

	users, err := svc.ManageUsers(testContext(), dmodel.DynamicFields{"id": "R1", composable.FieldM2mAdd: []any{"S1"}})
	require.NoError(t, err)
	require.Equal(t, 1, users.ClientErrors.Count())
	assert.Equal(t, ft.ErrorKey(errPrincipalNotUser, "iam"), users.ClientErrors[0].Key)

	services, err := svc.ManageServices(testContext(), dmodel.DynamicFields{
		"id": "R1", composable.FieldM2mAdd: []any{"U1"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, services.ClientErrors.Count())
	assert.Equal(t, ft.ErrorKey(errPrincipalNotService, "iam"), services.ClientErrors[0].Key)
}
