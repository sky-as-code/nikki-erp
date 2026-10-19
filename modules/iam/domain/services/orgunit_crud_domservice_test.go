package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func newOrgUnitService(units map[string]dmodel.DynamicFields, users map[string]dmodel.DynamicFields) (
	*OrgUnitCrudDomainServiceImpl, *stubCrudRepo,
) {
	base := newFakeCrudDomain()
	base.repo.rows = units
	userRepo := &stubCrudRepo{rows: users}
	svc := NewOrgUnitCrudDomainService(base, userRepo).(*OrgUnitCrudDomainServiceImpl)
	return svc, userRepo
}

func TestOrgUnitPathOfARootUnit(t *testing.T) {
	svc, _ := newOrgUnitService(nil, nil)
	unit := composable.NewDynamicEntityFrom(dmodel.DynamicFields{"id": "U1", domain.OrgUnitFieldOrgId: "O1"})
	vErrs := ft.ClientErrors{}

	require.NoError(t, svc.applyPath(testContext(), unit, nil, &vErrs))

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, []string{"O1", "U1"}, unit.GetFieldData().GetStrings(domain.OrgUnitFieldPath))
}

func TestOrgUnitPathExtendsTheParent(t *testing.T) {
	svc, _ := newOrgUnitService(map[string]dmodel.DynamicFields{
		"P1": {domain.OrgUnitFieldOrgId: "O1", domain.OrgUnitFieldPath: []string{"O1", "P1"}},
	}, nil)
	unit := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		"id": "U1", domain.OrgUnitFieldOrgId: "O1", domain.OrgUnitFieldParentId: "P1",
	})
	vErrs := ft.ClientErrors{}

	require.NoError(t, svc.applyPath(testContext(), unit, nil, &vErrs))

	assert.Equal(t, []string{"O1", "P1", "U1"}, unit.GetFieldData().GetStrings(domain.OrgUnitFieldPath))
}

func TestOrgUnitPathRefusesAParentOfAnotherOrg(t *testing.T) {
	svc, _ := newOrgUnitService(map[string]dmodel.DynamicFields{
		"P1": {domain.OrgUnitFieldOrgId: "O2", domain.OrgUnitFieldPath: []string{"O2", "P1"}},
	}, nil)
	unit := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		"id": "U1", domain.OrgUnitFieldOrgId: "O1", domain.OrgUnitFieldParentId: "P1",
	})
	vErrs := ft.ClientErrors{}

	require.NoError(t, svc.applyPath(testContext(), unit, nil, &vErrs))

	assert.Equal(t, 1, vErrs.Count())
}

func TestOrgUnitManageUsersMovesMembers(t *testing.T) {
	svc, userRepo := newOrgUnitService(map[string]dmodel.DynamicFields{"U1": {}}, map[string]dmodel.DynamicFields{
		"A": {}, "B": {domain.UserFieldOrgUnitId: "U1"}, "C": {domain.UserFieldOrgUnitId: "OTHER"},
	})

	result, err := svc.ManageUsers(testContext(), dmodel.DynamicFields{
		"id": "U1", composable.FieldM2mAdd: []any{"A"}, composable.FieldM2mRemove: []any{"B", "C"},
	})

	require.NoError(t, err)
	require.True(t, result.HasData)
	assert.Equal(t, 2, result.Data.AffectedCount, "C is in another unit and must be left alone")
	require.Len(t, userRepo.updates, 2)
	assert.Equal(t, "U1", userRepo.updates[0][domain.UserFieldOrgUnitId])
	assert.Nil(t, userRepo.updates[1][domain.UserFieldOrgUnitId])
}

func TestOrgUnitManageUsersRefusesAnUnknownUser(t *testing.T) {
	svc, userRepo := newOrgUnitService(map[string]dmodel.DynamicFields{"U1": {}}, map[string]dmodel.DynamicFields{})

	result, err := svc.ManageUsers(testContext(), dmodel.DynamicFields{
		"id": "U1", composable.FieldM2mAdd: []any{"MISSING"},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, userRepo.updates)
}
