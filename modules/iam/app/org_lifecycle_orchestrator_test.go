package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

type lifecycleBus struct {
	failing map[string]bool
}

func (this *lifecycleBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }
func (this *lifecycleBus) RequestNoReply(context.Context, cqrs.Request) error              { return nil }
func (this *lifecycleBus) IsRequestTypeRegistered(string) bool                            { return true }

func (this *lifecycleBus) Request(_ context.Context, request cqrs.Request, _ any) error {
	if this.failing[request.(*orglifecycle.OrgLifecycleCommand).TargetModule] {
		return errors.New("boom")
	}
	return nil
}

type statusCall struct {
	status  domain.OrgStatus
	message *string
	vars    map[string]any
}

type fakeOrgDomSvc struct {
	itOrg.OrgCrudDomainService
	statuses []statusCall
	deleted  bool
}

func (this *fakeOrgDomSvc) SetLifecycleStatus(
	_ corectx.Context, _ model.Id, status domain.OrgStatus, message *string, vars map[string]any,
) error {
	this.statuses = append(this.statuses, statusCall{status, message, vars})
	return nil
}

func (this *fakeOrgDomSvc) Delete(
	corectx.Context, itOrg.DeleteOrgCommand, ...composable.DeleteOptions,
) (*itOrg.DeleteOrgResult, error) {
	this.deleted = true
	return &itOrg.DeleteOrgResult{HasData: true}, nil
}

type fakeCatalog struct{ waves [][]string }

func (this fakeCatalog) ListModuleWaves(corectx.Context) ([][]string, error) { return this.waves, nil }

type fakeNotifier struct{ sent []itExt.NotifyUsersCommand }

func (this *fakeNotifier) NotifyUsers(_ corectx.Context, cmd itExt.NotifyUsersCommand) error {
	this.sent = append(this.sent, cmd)
	return nil
}

func newTestOrchestrator(failing ...string) (*OrgLifecycleOrchestrator, *fakeOrgDomSvc, *fakeNotifier) {
	failingSet := map[string]bool{}
	for _, module := range failing {
		failingSet[module] = true
	}
	domSvc := &fakeOrgDomSvc{}
	notifier := &fakeNotifier{}
	return &OrgLifecycleOrchestrator{
		domSvc:     domSvc,
		catalog:    fakeCatalog{waves: [][]string{{"essential"}, {"inventory", "sales"}}},
		notifier:   notifier,
		dispatcher: orglifecycle.NewDispatcher(&lifecycleBus{failing: failingSet}, nil),
		logger:     logging.NewLogger(),
	}, domSvc, notifier
}

func userContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{UserId: "USR-1"})
	return ctx
}

func TestInitOrg_AllModulesSucceed(t *testing.T) {
	orchestrator, domSvc, notifier := newTestOrchestrator()

	require.NoError(t, orchestrator.InitOrg(userContext(), "ORG-1"))

	require.Len(t, domSvc.statuses, 1)
	assert.Equal(t, domain.OrgStatusReady, domSvc.statuses[0].status)
	assert.Nil(t, domSvc.statuses[0].message)
	assert.Empty(t, notifier.sent)
}

func TestInitOrg_FailureKeepsProvisioningAndNotifies(t *testing.T) {
	orchestrator, domSvc, notifier := newTestOrchestrator("inventory")

	require.NoError(t, orchestrator.InitOrg(userContext(), "ORG-1"))

	require.Len(t, domSvc.statuses, 1)
	call := domSvc.statuses[0]
	assert.Equal(t, domain.OrgStatusProvisioning, call.status)
	assert.Equal(t, OrgStatusMsgInitFailed, *call.message)
	assert.Equal(t, "inventory", call.vars["failed_modules"])
	require.Len(t, notifier.sent, 1)
	assert.Equal(t, []model.Id{"USR-1"}, notifier.sent[0].RecipientUserIds)
	assert.Equal(t, "Fail to create organization", notifier.sent[0].Title)
}

func TestCleanUpOrg_SuccessRemovesTheOrg(t *testing.T) {
	orchestrator, domSvc, notifier := newTestOrchestrator()

	require.NoError(t, orchestrator.CleanUpOrg(userContext(), "ORG-1"))

	assert.True(t, domSvc.deleted)
	assert.Empty(t, domSvc.statuses)
	assert.Empty(t, notifier.sent)
}

func TestCleanUpOrg_FailureKeepsTheOrgDeleted(t *testing.T) {
	orchestrator, domSvc, notifier := newTestOrchestrator("sales")

	require.NoError(t, orchestrator.CleanUpOrg(userContext(), "ORG-1"))

	assert.False(t, domSvc.deleted)
	require.Len(t, domSvc.statuses, 1)
	assert.Equal(t, domain.OrgStatusDeleted, domSvc.statuses[0].status)
	assert.Equal(t, OrgStatusMsgCleanupFailed, *domSvc.statuses[0].message)
	assert.Equal(t, "essential", domSvc.statuses[0].vars["skipped_modules"],
		"cleanup runs in reverse: essential waits for sales and is skipped")
	assert.Len(t, notifier.sent, 1)
}
