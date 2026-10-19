package orglifecycle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type recordingService struct {
	action string
	orgId  model.Id
	ctx    corectx.Context
	err    error
}

func (this *recordingService) InitOrg(ctx corectx.Context, orgId model.Id) error {
	this.action, this.orgId, this.ctx = ActionInitOrg, orgId, ctx
	return this.err
}

func (this *recordingService) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	this.action, this.orgId, this.ctx = ActionCleanUpOrg, orgId, ctx
	return this.err
}

func salesCommand(action string) *OrgLifecycleCommand {
	return &OrgLifecycleCommand{
		RequestId: "REQ-1", TargetModule: "sales", Action: action, OrgId: "ORG-1", RequestedBy: "USR-1",
	}
}

func TestHandler_RoutesActionAndActsAsService(t *testing.T) {
	for _, action := range []string{ActionInitOrg, ActionCleanUpOrg} {
		svc := &recordingService{}
		handler := &lifecycleHandler{moduleName: "sales", svc: svc}

		reply, err := handler.answer(corectx.NewRequestContext(context.Background()), salesCommand(action))

		require.NoError(t, err)
		assert.Equal(t, "sales", reply.Result.Module)
		assert.Equal(t, "REQ-1", reply.Result.RequestId)
		assert.Equal(t, action, svc.action)
		assert.Equal(t, model.Id("ORG-1"), svc.orgId)

		perms := svc.ctx.GetPermissions()
		assert.Equal(t, corectx.PrincipalKindService, perms.Principal.Kind)
		assert.Equal(t, model.Id("ORG-1"), *perms.Principal.OrgId)
		assert.Equal(t, model.Id("USR-1"), perms.UserId, "rows are stamped with the requesting user")
	}
}

func TestHandler_KeepsDomainConstraints(t *testing.T) {
	svc := &recordingService{}
	handler := &lifecycleHandler{moduleName: "sales", svc: svc}
	incoming := corectx.NewRequestContextF(context.Background(), "core", map[string]any{"tenant_id": "T-1"})

	_, err := handler.answer(incoming, salesCommand(ActionInitOrg))

	require.NoError(t, err)
	assert.Equal(t, "T-1", svc.ctx.GetDomainConstraints()["tenant_id"])
}

func TestHandler_ServiceErrorBecomesAnError(t *testing.T) {
	handler := &lifecycleHandler{moduleName: "sales", svc: &recordingService{err: errors.New("boom")}}

	_, err := handler.answer(context.Background(), salesCommand(ActionInitOrg))

	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}

func TestHandler_RefusesMisaddressedOrIncompleteCommands(t *testing.T) {
	handler := &lifecycleHandler{moduleName: "sales", svc: &recordingService{}}

	wrongModule := salesCommand(ActionInitOrg)
	wrongModule.TargetModule = "inventory"
	unknownAction := salesCommand("dropOrg")
	noOrg := salesCommand(ActionCleanUpOrg)
	noOrg.OrgId = ""

	for _, cmd := range []*OrgLifecycleCommand{wrongModule, unknownAction, noOrg} {
		_, err := handler.answer(context.Background(), cmd)
		assert.Error(t, err)
	}
}
