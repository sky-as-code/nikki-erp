package services

import (
	"slices"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

var (
	ErrKeyOrgDeleted    = ft.ErrorKey("err_org_deleted", c.IamModuleName)
	ErrKeyOrgNotDeleted = ft.ErrorKey("err_org_not_deleted", c.IamModuleName)
)

// asProvisioning makes every new org start in provisioning, whatever the caller sent: only the
// org lifecycle moves it on.
func asProvisioning(cmd itOrg.CreateOrgCommand) itOrg.CreateOrgCommand {
	out := make(dmodel.DynamicFields, len(cmd))
	for key, value := range cmd {
		if !slices.Contains(domain.OrgStatusFields, key) {
			out[key] = value
		}
	}
	out[domain.OrgFieldStatus] = string(domain.OrgStatusProvisioning)
	return out
}

// Update refuses to change a deleted org, except for its status fields, which the org lifecycle
// still writes while modules clean up.
func (this *OrgCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd itOrg.UpdateOrgCommand, options ...composable.UpdateOptions,
) (*itOrg.UpdateOrgResult, error) {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(assertOnlyStatusIfDeleted, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

func assertOnlyStatusIfDeleted(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found == nil || orgStatusOf(found.GetFieldData()) != domain.OrgStatusDeleted {
		return nil
	}
	constraints := ctx.GetDomainConstraints()
	for key := range input.GetFieldData() {
		if _, isConstraint := constraints[key]; isConstraint {
			continue
		}
		if isBookkeepingField(key) || slices.Contains(domain.OrgStatusFields, key) {
			continue
		}
		vErrs.Append(*orgDeletedViolation())
		return nil
	}
	return nil
}

// isBookkeepingField names what every update carries or has injected, which changes nothing a
// user set.
func isBookkeepingField(key string) bool {
	switch key {
	case basemodel.FieldId, basemodel.FieldEtag, basemodel.FieldUpdatedAt, basemodel.FieldUpdatedBy:
		return true
	}
	return false
}

// SetArchived refuses a deleted org: it was archived as part of its deletion, and unarchiving it
// would bring back an org whose data modules are removing.
func (this *OrgCrudDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd itOrg.SetOrgArchivedCommand,
) (*itOrg.SetOrgArchivedResult, error) {
	status, cErrs, err := this.loadStatus(ctx, cmd)
	if err != nil || cErrs != nil {
		return &itOrg.SetOrgArchivedResult{ClientErrors: util.ValueOrZeroOf(cErrs)}, err
	}
	if status == domain.OrgStatusDeleted {
		return &itOrg.SetOrgArchivedResult{ClientErrors: ft.ClientErrors{*orgDeletedViolation()}}, nil
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}

// assertDeletedBeforeRemoval allows the row to go only once the org lifecycle has marked the org
// deleted and every module has cleaned up; a client's delete marks it, it never removes it.
func assertDeletedBeforeRemoval(
	_ corectx.Context, _ *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found != nil && orgStatusOf(found.GetFieldData()) != domain.OrgStatusDeleted {
		vErrs.Append(*ft.NewBusinessViolation(domain.OrgFieldStatus, ErrKeyOrgNotDeleted,
			"the organization must be deleted and cleaned up before it is removed"))
	}
	return nil
}

func (this *OrgCrudDomainServiceImpl) loadStatus(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (domain.OrgStatus, *ft.ClientErrors, error) {
	found, cErrs, err := this.loadOrg(ctx, composable.ParamId(keys))
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	return orgStatusOf(found), nil, nil
}

func orgStatusOf(fields dmodel.DynamicFields) domain.OrgStatus {
	return util.ValueOrZeroOf(domain.NewOrganizationFrom(fields).GetStatus())
}

func orgDeletedViolation() *ft.ClientErrorItem {
	return ft.NewBusinessViolation(domain.OrgFieldStatus, ErrKeyOrgDeleted,
		"the organization is being deleted and can no longer be changed")
}

// MarkDeleted starts an org's deletion: status deleted and archived in one write, so the org is
// never seen deleted but still active. Marking an org already deleted succeeds again, which is
// how a failed cleanup is retried.
func (this *OrgCrudDomainServiceImpl) MarkDeleted(
	ctx corectx.Context, cmd itOrg.DeleteOrgCommand,
) (*itOrg.DeleteOrgResult, error) {
	found, cErrs, err := this.loadOrg(ctx, composable.ParamId(cmd))
	if err != nil || cErrs != nil {
		return &itOrg.DeleteOrgResult{ClientErrors: util.ValueOrZeroOf(cErrs)}, err
	}
	result, err := corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       "mark organization deleted",
		DbRepoGetter: this.Repository(),
		Data: dmodel.DynamicFields{
			basemodel.FieldId:                found[basemodel.FieldId],
			basemodel.FieldEtag:              found[basemodel.FieldEtag],
			basemodel.FieldIsArchived:        true,
			domain.OrgFieldStatus:            string(domain.OrgStatusDeleted),
			domain.OrgFieldStatusMessage:     nil,
			domain.OrgFieldStatusMessageVars: nil,
		},
	})
	return result, errors.Wrap(err, "mark organization deleted")
}

// SetLifecycleStatus records what the org lifecycle concluded. message and vars are cleared when
// nil.
func (this *OrgCrudDomainServiceImpl) SetLifecycleStatus(
	ctx corectx.Context, orgId model.Id, status domain.OrgStatus, message *string, vars map[string]any,
) error {
	found, cErrs, err := this.loadOrg(ctx, orgId)
	if err != nil {
		return err
	}
	if cErrs != nil {
		return cErrs.ToError()
	}
	update := domain.NewOrganization()
	update.SetId(&orgId)
	update.SetEtag(domain.NewOrganizationFrom(found).GetEtag())
	update.SetStatus(&status)
	update.SetStatusMessage(message)
	update.SetStatusMessageVars(vars)

	result, err := this.Update(ctx, update.GetFieldData())
	if err != nil {
		return errors.Wrap(err, "set organization lifecycle status")
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "set organization lifecycle status")
	}
	return nil
}

func (this *OrgCrudDomainServiceImpl) loadOrg(
	ctx corectx.Context, orgId model.Id,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	result, err := this.GetById(ctx, dmodel.DynamicFields{basemodel.FieldId: orgId})
	if err != nil {
		return nil, nil, errors.Wrap(err, "load organization")
	}
	if result.ClientErrors.Count() > 0 {
		return nil, &result.ClientErrors, nil
	}
	if !result.HasData {
		return nil, &ft.ClientErrors{*ft.NewNotFoundError(basemodel.FieldId)}, nil
	}
	return result.Data.Item, nil, nil
}
