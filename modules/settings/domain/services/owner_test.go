package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

const (
	unitOrgId   = model.Id("01ORGA0000000000000000000A")
	chosenOrgId = model.Id("01ORGB0000000000000000000B")
)

// userInTwoOrgs is the reported case: a member of two orgs whose org unit sits in the first.
func userInTwoOrgs() corectx.Context {
	unitOrg := unitOrgId
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("01USER0000000000000000001")},
		UserOrgIds:   ds.NewSetFrom(unitOrgId, chosenOrgId),
		OrgUnitOrgId: &unitOrg,
	})
	return ctx
}

func TestActingOrgId_AuthorizedOrgWinsOverTheOrgUnit(t *testing.T) {
	ctx := userInTwoOrgs()
	dyn.WithActingOrg(ctx, chosenOrgId)

	assert.Equal(t, string(chosenOrgId), actingOrgId(ctx))
}

func TestActingOrgId_WithoutActingOrgFallsBackToTheOrgUnit(t *testing.T) {
	assert.Equal(t, string(unitOrgId), actingOrgId(userInTwoOrgs()))
}

func TestActingOrgId_WithoutActingOrgFallsBackToTheOnlyMembership(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{UserOrgIds: ds.NewSetFrom(chosenOrgId)})

	assert.Equal(t, string(chosenOrgId), actingOrgId(ctx))
}

func TestActingOrgId_ServicePrincipalKeepsItsOrg(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	orgId := chosenOrgId
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{Kind: corectx.PrincipalKindService, OrgId: &orgId},
	})

	assert.Equal(t, string(chosenOrgId), actingOrgId(ctx))
}
