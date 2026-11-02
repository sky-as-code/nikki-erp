package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

const (
	orgOfUnit   = model.Id("01ORGA0000000000000000000A")
	orgChosen   = model.Id("01ORGB0000000000000000000B")
	orgStranger = model.Id("01ORGC0000000000000000000C")
)

// fakeSettingsDomain records the acting org each call saw; the rest of the interface is unused.
type fakeSettingsDomain struct {
	it.SettingsDomainService
	calls      int
	actingOrgs []model.Id
}

func (this *fakeSettingsDomain) record(ctx corectx.Context) {
	this.calls++
	orgId, _ := dyn.ActingOrgOf(ctx)
	this.actingOrgs = append(this.actingOrgs, orgId)
}

func (this *fakeSettingsDomain) GetSettings(
	ctx corectx.Context, _ string, _ string, _ it.GetSettingsQuery,
) (*it.GetSettingsResult, error) {
	this.record(ctx)
	return &it.GetSettingsResult{HasData: true}, nil
}

func (this *fakeSettingsDomain) SetSettings(
	ctx corectx.Context, _ string, _ string, _ it.SetSettingsCommand,
) (*it.SetSettingsResult, error) {
	this.record(ctx)
	return &it.SetSettingsResult{HasData: true}, nil
}

// orgMemberCaller belongs to orgOfUnit (where its org unit sits) and orgChosen, holding the given
// grants on settings_record.
func orgMemberCaller(entitlements ...string) corectx.Context {
	unitOrg := orgOfUnit
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("01USER0000000000000000001")},
		UserOrgIds:   ds.NewSetFrom(orgOfUnit, orgChosen),
		OrgUnitOrgId: &unitOrg,
		Entitlements: ds.NewSetFrom(entitlements...),
	})
	return ctx
}

func bareOrgGrant(action string) string {
	return reguard.BuildExpression(action, settingsRecordResource, reguard.ResourceScopeOrg, nil)
}

func tenantGrant(action string) string {
	return reguard.BuildExpression(action, settingsRecordResource, reguard.ResourceScopeTenant, nil)
}

func TestGetOrgSettings_NamedOrgBecomesTheActingOrg(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)
	orgId := orgChosen

	result, err := svc.GetOrgSettings(orgMemberCaller(bareOrgGrant(actionRead)),
		it.GetSettingsQuery{ModuleKey: "vendingmachine", OrgId: &orgId})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count(), "a bare org grant answers in an org the caller belongs to")
	assert.Equal(t, []model.Id{orgChosen}, domain.actingOrgs)
}

func TestGetOrgSettings_RefusesAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)
	orgId := orgStranger

	result, err := svc.GetOrgSettings(orgMemberCaller(tenantGrant(actionRead)),
		it.GetSettingsQuery{ModuleKey: "vendingmachine", OrgId: &orgId})

	require.NoError(t, err, "a foreign org is the caller's mistake, not a server fault")
	require.Positive(t, result.ClientErrors.Count())
	assert.Equal(t, "org_id", result.ClientErrors[0].Field)
	assert.Zero(t, domain.calls, "even a tenant grant must not reach an org the caller is not in")
}

func TestGetOrgSettings_WithoutOrgIdLeavesTheOrgToTheDomain(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)

	result, err := svc.GetOrgSettings(orgMemberCaller(tenantGrant(actionRead)),
		it.GetSettingsQuery{ModuleKey: "vendingmachine"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, []model.Id{""}, domain.actingOrgs, "no org_id must not invent an acting org")
}

func TestSetOrgSettings_NamedOrgBecomesTheActingOrg(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)
	orgId := orgChosen

	result, err := svc.SetOrgSettings(orgMemberCaller(bareOrgGrant(actionUpdate)),
		it.SetSettingsCommand{ModuleKey: "vendingmachine", OrgId: &orgId})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, []model.Id{orgChosen}, domain.actingOrgs)
}

func TestSetOrgSettings_RefusesAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)
	orgId := orgStranger

	result, err := svc.SetOrgSettings(orgMemberCaller(tenantGrant(actionUpdate)),
		it.SetSettingsCommand{ModuleKey: "vendingmachine", OrgId: &orgId})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domain.calls)
}

// Without org_id the check stays at tenant scope: a bare org grant never answered it before.
func TestSetOrgSettings_WithoutOrgIdKeepsTheTenantScopeCheck(t *testing.T) {
	domain := &fakeSettingsDomain{}
	svc := NewOrgSettingsAppServiceImpl(domain)

	result, err := svc.SetOrgSettings(orgMemberCaller(bareOrgGrant(actionUpdate)),
		it.SetSettingsCommand{ModuleKey: "vendingmachine"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domain.calls)
}
