package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
)

const testCreatorId = model.Id("01JWNZ3XVWDP73JGHRRBFAHQY5")

type fakeOrgSettingsInit struct {
	initialized []model.Id
}

func (this *fakeOrgSettingsInit) InitOrgSettings(
	_ corectx.Context, cmd itExt.InitOwnerSettingsCommand,
) (*itExt.InitOwnerSettingsResult, error) {
	this.initialized = append(this.initialized, cmd.OwnerId)
	return &itExt.InitOwnerSettingsResult{HasData: true}, nil
}

func newOrgTestService() (*OrgCrudDomainServiceImpl, *fakeCrudDomain, *fakeOrgSettingsInit) {
	base := newFakeCrudDomain()
	settings := &fakeOrgSettingsInit{}
	svc := NewOrgCrudDomainService(base, settings, nil).(*OrgCrudDomainServiceImpl)
	return svc, base, settings
}

func orgTestContext(principal corectx.Principal) corectx.Context {
	ctx := testContext()
	ctx.SetPermissions(corectx.ContextPermissions{UserId: principal.Id, Principal: principal})
	return ctx
}

func orgCreateCommand() dmodel.DynamicFields {
	return dmodel.DynamicFields{domain.OrgFieldDisplayName: "Acme"}
}

func TestOrgCreate_UserCreatorBecomesMember(t *testing.T) {
	svc, base, settings := newOrgTestService()
	ctx := orgTestContext(corectx.Principal{Kind: corectx.PrincipalKindUser, Id: testCreatorId})

	result, err := svc.Create(ctx, orgCreateCommand())

	require.NoError(t, err)
	require.True(t, result.HasData)
	orgId := model.Id(base.created["id"].(string))
	assert.Equal(t, []model.Id{orgId}, settings.initialized)
	require.Len(t, base.repo.base.m2mCalls, 1)
	call := base.repo.base.m2mCalls[0]
	assert.Equal(t, orgId, call.SrcId)
	assert.Equal(t, domain.UserSchemaName, call.DestSchemaName)
	assert.Equal(t, []model.Id{testCreatorId}, call.AssociatedIds.ToSlice())
	assert.True(t, base.repo.tranx.committed)
}

func TestOrgCreate_NonUserPrincipalIsNotAdded(t *testing.T) {
	principals := map[string]corectx.Principal{
		"zero":    {},
		"service": {Kind: corectx.PrincipalKindService, Id: testCreatorId},
	}
	for name, principal := range principals {
		svc, base, settings := newOrgTestService()

		result, err := svc.Create(orgTestContext(principal), orgCreateCommand())

		require.NoError(t, err, name)
		require.True(t, result.HasData, name)
		assert.Len(t, settings.initialized, 1, name)
		assert.Nil(t, base.repo.base, name)
		assert.True(t, base.repo.tranx.committed, name)
	}
}

func TestOrgCreate_RefusedMembershipRollsBackTheCreate(t *testing.T) {
	svc, base, _ := newOrgTestService()
	base.repo.base = &stubBaseRepo{
		schema:     base.repo.schema,
		m2mRefusal: ft.NewBusinessViolation("add", "err_test", "refused"),
	}
	ctx := orgTestContext(corectx.Principal{Kind: corectx.PrincipalKindUser, Id: testCreatorId})

	result, err := svc.Create(ctx, orgCreateCommand())

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Len(t, base.repo.base.m2mCalls, 1)
	assert.False(t, base.repo.tranx.committed)
	assert.True(t, base.repo.tranx.rolledBack)
}
