package app

import (
	"slices"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// NewOrgCrudApplicationService is handed the composable default by the iam_org onion.
func NewOrgCrudApplicationService(
	base composable.CrudApplicationService, domSvc itOrg.OrgCrudDomainService, lifecycle *OrgLifecycleOrchestrator,
) itOrg.OrgCrudApplicationService {
	return &OrgCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc, lifecycle: lifecycle}
}

type OrgCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc    itOrg.OrgCrudDomainService
	lifecycle *OrgLifecycleOrchestrator
}

// Create answers with the org still provisioning; every module's defaults are written after.
func (this *OrgCrudApplicationServiceImpl) Create(
	ctx corectx.Context, cmd itOrg.CreateOrgCommand,
) (*itOrg.CreateOrgResult, error) {
	result, err := this.CrudApplicationService.Create(ctx, withoutStatusFields(cmd))
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	this.lifecycle.StartInit(ctx, composable.ParamId(result.Data))
	return result, nil
}

func (this *OrgCrudApplicationServiceImpl) Update(
	ctx corectx.Context, cmd itOrg.UpdateOrgCommand,
) (*itOrg.UpdateOrgResult, error) {
	return this.CrudApplicationService.Update(ctx, withoutStatusFields(cmd))
}

// Delete marks the org deleted and archived and answers; modules clean up after, and the row is
// removed once they all have. Deleting an org already marked deleted retries the cleanup.
func (this *OrgCrudApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd itOrg.DeleteOrgCommand,
) (*itOrg.DeleteOrgResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd,
		composable.ActionPermission{ActionCode: composable.PermissionDelete})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	result, err := this.domSvc.MarkDeleted(ctx, cmd)
	if err != nil || result.ClientErrors.Count() > 0 {
		return result, err
	}
	// Answered as a delete, which reports no etag: the client is done with this record.
	result.Data.Etag = ""
	this.lifecycle.StartCleanUp(ctx, composable.ParamId(cmd))
	return result, nil
}

func (this *OrgCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrg.ManageOrgUsersCommand,
) (*itOrg.ManageOrgUsersResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{ActionCode: actionManageUsers})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

// withoutStatusFields drops what only the org lifecycle writes.
func withoutStatusFields(cmd dmodel.DynamicFields) dmodel.DynamicFields {
	out := make(dmodel.DynamicFields, len(cmd))
	for key, value := range cmd {
		if !slices.Contains(domain.OrgStatusFields, key) {
			out[key] = value
		}
	}
	return out
}
