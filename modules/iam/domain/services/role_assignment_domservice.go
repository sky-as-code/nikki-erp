package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// rolePrincipalKind is one kind of principal a role is assigned to. The junction rows are the
// same ones the principal-side actions write; only the end they are managed from differs.
type rolePrincipalKind struct {
	action        string
	destSchema    string
	assignIdField string
	addReason     domain.PermissionHistoryReason
	removeReason  domain.PermissionHistoryReason
	// refusedType, when set, is the principal type the added ids must not be: users and services
	// share the junction, so the schema cannot tell them apart.
	refusedType *principalTypeScope
}

// ManageUsers assigns users to, and removes them from, one role.
func (this *RoleCrudDomainServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itRole.ManageRoleUsersCommand,
) (*itRole.ManageRoleUsersResult, error) {
	return this.managePrincipals(ctx, cmd, rolePrincipalKind{
		action:        "manage role users",
		destSchema:    domain.UserSchemaName,
		assignIdField: domain.RoleUserAssignFieldId,
		addReason:     domain.PermissionHistoryReasonRoleAdded,
		removeReason:  domain.PermissionHistoryReasonRoleRemoved,
		refusedType:   &serviceTypeScope,
	})
}

// ManageServices assigns services to, and removes them from, one role: the same junction rows as
// ManageUsers, refusing user ids instead of service ids.
func (this *RoleCrudDomainServiceImpl) ManageServices(
	ctx corectx.Context, cmd itRole.ManageRoleServicesCommand,
) (*itRole.ManageRoleServicesResult, error) {
	return this.managePrincipals(ctx, cmd, rolePrincipalKind{
		action:        "manage role services",
		destSchema:    domain.ServiceSchemaName,
		assignIdField: domain.RoleUserAssignFieldId,
		addReason:     domain.PermissionHistoryReasonRoleAdded,
		removeReason:  domain.PermissionHistoryReasonRoleRemoved,
		refusedType:   &userTypeScope,
	})
}

// ManageGroups assigns groups to, and removes them from, one role.
func (this *RoleCrudDomainServiceImpl) ManageGroups(
	ctx corectx.Context, cmd itRole.ManageRoleGroupsCommand,
) (*itRole.ManageRoleGroupsResult, error) {
	return this.managePrincipals(ctx, cmd, rolePrincipalKind{
		action:        "manage role groups",
		destSchema:    domain.GroupSchemaName,
		assignIdField: domain.RoleGroupAssignFieldId,
		addReason:     domain.PermissionHistoryReasonRoleAddedGroup,
		removeReason:  domain.PermissionHistoryReasonRoleRemovedGroup,
	})
}

// managePrincipals writes the junction rows, the permission rebuild and the audit trail in one
// transaction. One rebuild covers both directions: it drops every cached row this role's
// entitlements produced and recomputes them, so removed principals lose them and added ones gain
// them. Group audit rows name the group, not its members.
func (this *RoleCrudDomainServiceImpl) managePrincipals(
	ctx corectx.Context, cmd composable.ManageM2mCommand, kind rolePrincipalKind,
) (*composable.MutateResult, error) {
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          kind.action,
		DestSchemaName:  kind.destSchema,
		JunctionIdField: kind.assignIdField,
		Validate: func(ctx corectx.Context, change composable.M2mChange, vErrs *ft.ClientErrors) error {
			if err := this.assertRoleAssignable(ctx, change.SrcId, vErrs); err != nil || kind.refusedType == nil {
				return err
			}
			return kind.refusedType.refuseIdsOfType(ctx, this.userRepo, composable.FieldM2mAdd, change.Added, vErrs)
		},
		AfterChange: func(tranxCtx corectx.Context, change composable.M2mChange) error {
			if err := this.permRepo.RebuildUserPermissionsForRole(tranxCtx, change.SrcId); err != nil {
				return err
			}
			err := this.auditor.recordRoleTransition(
				tranxCtx, domain.PermissionHistoryEffectGrant, kind.addReason, change.SrcId, change.Added)
			if err != nil {
				return err
			}
			return this.auditor.recordRoleTransition(
				tranxCtx, domain.PermissionHistoryEffectRevoke, kind.removeReason, change.SrcId, change.Removed)
		},
	})
}

// assertRoleAssignable refuses a system role: it is held by every user of the tenant without
// an assignment row, so an assignment would be meaningless and a removal impossible.
func (this *RoleCrudDomainServiceImpl) assertRoleAssignable(
	ctx corectx.Context, roleId model.Id, vErrs *ft.ClientErrors,
) error {
	found, err := this.roleRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: string(roleId)},
		Fields: []string{domain.RoleFieldIsSystem},
	})
	if err != nil {
		return err
	}
	if found.ClientErrors.Count() > 0 {
		vErrs.Concat(found.ClientErrors)
		return nil
	}
	if !found.HasData {
		vErrs.Append(*ft.NewNotFoundError(basemodel.FieldId))
		return nil
	}
	isSystem := found.Data.GetBool(domain.RoleFieldIsSystem)
	if isSystem != nil && *isSystem {
		vErrs.Append(*ft.NewBusinessViolation(
			basemodel.FieldId, ft.ErrorKey("err_system_role_assignment_not_allowed", "iam"),
			"a system role is held by every user and cannot be assigned or removed",
		))
	}
	return nil
}
