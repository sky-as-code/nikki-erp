package app

import (
	"context"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// The status_message keys, translated by the frontend with status_message_vars.
const (
	OrgStatusMsgInitFailed    = "iam:org_status.init_failed"
	OrgStatusMsgCleanupFailed = "iam:org_status.cleanup_failed"
)

type OrgLifecycleOrchestratorParam struct {
	DomSvc     itOrg.OrgCrudDomainService
	Catalog    itExt.ModuleCatalogExtService
	Notifier   itExt.OrgNotificationExtService
	Dispatcher *orglifecycle.Dispatcher
	Config     config.ConfigService
	Logger     logging.LoggerService
}

func NewOrgLifecycleOrchestrator(param OrgLifecycleOrchestratorParam) *OrgLifecycleOrchestrator {
	timeoutSecs := param.Config.GetInt(c.OrgLifecycleModuleTimeoutSecs, int(orglifecycle.DefaultModuleTimeout.Seconds()))
	return &OrgLifecycleOrchestrator{
		domSvc:     param.DomSvc,
		catalog:    param.Catalog,
		notifier:   param.Notifier,
		dispatcher: param.Dispatcher.WithTimeout(time.Duration(timeoutSecs) * time.Second),
		logger:     param.Logger,
	}
}

// OrgLifecycleOrchestrator asks every installed module to set up a new org or clean up a deleted
// one, and records the outcome on the org.
//
// It runs in this process only: if the process stops mid-way, the org stays provisioning or
// deleted with nothing resuming it (docs/problems/iam/001).
type OrgLifecycleOrchestrator struct {
	domSvc     itOrg.OrgCrudDomainService
	catalog    itExt.ModuleCatalogExtService
	notifier   itExt.OrgNotificationExtService
	dispatcher *orglifecycle.Dispatcher
	logger     logging.LoggerService
}

// StartInit runs InitOrg off the request, which has already answered with the org provisioning.
func (this *OrgLifecycleOrchestrator) StartInit(ctx corectx.Context, orgId model.Id) {
	this.runDetached(ctx, "init", orgId, this.InitOrg)
}

// StartCleanUp runs CleanUpOrg off the request, which has already answered with the org deleted.
func (this *OrgLifecycleOrchestrator) StartCleanUp(ctx corectx.Context, orgId model.Id) {
	this.runDetached(ctx, "clean up", orgId, this.CleanUpOrg)
}

// runDetached keeps the request's identity and tenant but not its cancellation or transaction:
// the request ends long before the modules answer.
func (this *OrgLifecycleOrchestrator) runDetached(
	ctx corectx.Context, verb string, orgId model.Id, run func(corectx.Context, model.Id) error,
) {
	bgCtx := corectx.CloneWithInner(ctx, context.WithoutCancel(ctx.InnerContext()))
	bgCtx.SetDbTranx(nil)

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				this.logger.Errorf("iam: org lifecycle %s of org %s panicked: %v", verb, orgId, recovered)
			}
		}()
		if err := run(bgCtx, orgId); err != nil {
			this.logger.Errorf("iam: org lifecycle %s of org %s: %v", verb, orgId, err)
		}
	}()
}

// InitOrg asks every module to write its defaults, then marks the org ready, or records which
// modules failed and tells the creator.
func (this *OrgLifecycleOrchestrator) InitOrg(ctx corectx.Context, orgId model.Id) error {
	failed, skipped, requestId, err := this.dispatch(ctx, orglifecycle.ActionInitOrg, orgId, false)
	if err != nil {
		return err
	}
	if len(failed) == 0 {
		return this.domSvc.SetLifecycleStatus(ctx, orgId, domain.OrgStatusReady, nil, nil)
	}

	this.recordFailure(ctx, orgId, domain.OrgStatusProvisioning, OrgStatusMsgInitFailed, failed, skipped)
	this.notify(ctx, orgId, "iam-org-init-"+requestId, "Fail to create organization",
		"Failed to set up defaults for modules: "+strings.Join(failed, ", "))
	return nil
}

// CleanUpOrg asks every module to delete the org's data, in reverse dependency order, then
// removes the org row.
func (this *OrgLifecycleOrchestrator) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	failed, skipped, requestId, err := this.dispatch(ctx, orglifecycle.ActionCleanUpOrg, orgId, true)
	if err != nil {
		return err
	}
	if len(failed) == 0 {
		failed, err = this.removeOrg(ctx, orgId)
		if err != nil || len(failed) == 0 {
			return err
		}
	}

	this.recordFailure(ctx, orgId, domain.OrgStatusDeleted, OrgStatusMsgCleanupFailed, failed, skipped)
	this.notify(ctx, orgId, "iam-org-cleanup-"+requestId, "Fail to delete organization",
		"Failed to clean up modules: "+strings.Join(failed, ", "))
	return nil
}

// dispatch returns the modules that failed and those skipped after a failure. A module list that
// cannot be read fails the whole action under Essential's name, since nobody was asked.
func (this *OrgLifecycleOrchestrator) dispatch(
	ctx corectx.Context, action string, orgId model.Id, reverse bool,
) (failed []string, skipped []string, requestId string, err error) {
	waves, err := this.catalog.ListModuleWaves(ctx)
	if err != nil {
		this.logger.Errorf("iam: org lifecycle %s of org %s: listing modules: %v", action, orgId, err)
		return []string{"essential"}, nil, string(orgId), nil
	}
	if reverse {
		waves = orglifecycle.ReverseWaves(waves)
	}

	outcome, err := this.dispatcher.Dispatch(ctx, action, waves, orgId, ctx.GetPermissions().UserId)
	if err != nil {
		return nil, nil, "", errors.Wrapf(err, "dispatch %s", action)
	}
	return outcome.FailedModules(), outcome.SkippedModules(), outcome.RequestId, nil
}

// removeOrg deletes the org row once every module has cleaned up. A refusal is reported as an
// iam failure: something still references the org.
func (this *OrgLifecycleOrchestrator) removeOrg(ctx corectx.Context, orgId model.Id) ([]string, error) {
	org := domain.NewOrganization()
	org.SetId(&orgId)
	result, err := this.domSvc.Delete(ctx, org.GetFieldData())
	if err != nil {
		return nil, errors.Wrap(err, "remove cleaned-up organization")
	}
	if result.ClientErrors.Count() > 0 {
		this.logger.Errorf("iam: removing org %s after cleanup was refused: %v", orgId, result.ClientErrors.ToError())
		return []string{c.IamModuleName}, nil
	}
	return nil, nil
}

func (this *OrgLifecycleOrchestrator) recordFailure(
	ctx corectx.Context, orgId model.Id, status domain.OrgStatus, messageKey string, failed []string, skipped []string,
) {
	vars := map[string]any{
		"failed_modules":  strings.Join(failed, ", "),
		"skipped_modules": strings.Join(skipped, ", "),
	}
	if err := this.domSvc.SetLifecycleStatus(ctx, orgId, status, &messageKey, vars); err != nil {
		this.logger.Errorf("iam: recording lifecycle failure of org %s: %v", orgId, err)
	}
}

// notify tells the user who started the action. A failed notification is only logged: the org's
// status_message already carries the outcome.
func (this *OrgLifecycleOrchestrator) notify(
	ctx corectx.Context, orgId model.Id, commandId string, title string, message string,
) {
	userId := ctx.GetPermissions().UserId
	if userId == "" {
		return
	}
	err := this.notifier.NotifyUsers(ctx, itExt.NotifyUsersCommand{
		CommandId:        commandId,
		OrgId:            orgId,
		RecipientUserIds: []model.Id{userId},
		Title:            title,
		Message:          message,
		Severity:         "danger",
		Metadata:         map[string]any{"org_id": string(orgId)},
	})
	if err != nil {
		this.logger.Errorf("iam: notifying user %s about org %s: %v", userId, orgId, err)
	}
}
