package services

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// rolePrincipalChange is one role-side assignment call, for either kind of principal. The
// junction rows are the same ones the principal-side endpoints write; only the end they are
// managed from differs.
type rolePrincipalChange struct {
	action        string
	destSchema    string
	assignIdField string
	addReason     domain.PermissionHistoryReason
	removeReason  domain.PermissionHistoryReason
	roleId        model.Id
	add           datastructure.Set[model.Id]
	remove        datastructure.Set[model.Id]
}

// ManageRoleUsers assigns users to, and removes them from, one role.
func (this *RoleDomainServiceImpl) ManageRoleUsers(
	ctx corectx.Context, cmd itRole.ManageRoleUsersCommand,
) (*itRole.ManageRoleUsersResult, error) {
	return this.manageRolePrincipals(ctx, rolePrincipalChange{
		action:        "manage role users",
		destSchema:    domain.UserSchemaName,
		assignIdField: domain.RoleUserAssignFieldId,
		addReason:     domain.PermissionHistoryReasonRoleAdded,
		removeReason:  domain.PermissionHistoryReasonRoleRemoved,
		roleId:        cmd.RoleId,
		add:           cmd.Add,
		remove:        cmd.Remove,
	})
}

// ManageRoleGroups assigns groups to, and removes them from, one role.
func (this *RoleDomainServiceImpl) ManageRoleGroups(
	ctx corectx.Context, cmd itRole.ManageRoleGroupsCommand,
) (*itRole.ManageRoleGroupsResult, error) {
	return this.manageRolePrincipals(ctx, rolePrincipalChange{
		action:        "manage role groups",
		destSchema:    domain.GroupSchemaName,
		assignIdField: domain.RoleGroupAssignFieldId,
		addReason:     domain.PermissionHistoryReasonRoleAddedGroup,
		removeReason:  domain.PermissionHistoryReasonRoleRemovedGroup,
		roleId:        cmd.RoleId,
		add:           cmd.Add,
		remove:        cmd.Remove,
	})
}

// manageRolePrincipals expects the caller's transaction: the junction rows, the permission
// rebuild and the audit trail must commit together.
func (this *RoleDomainServiceImpl) manageRolePrincipals(
	ctx corectx.Context, change rolePrincipalChange,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	cErrs, err := this.assertRoleAssignable(ctx, change.roleId)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}
	result, err := corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             change.action,
		DbRepoGetter:       this.roleRepo,
		DestSchemaName:     change.destSchema,
		SrcId:              change.roleId,
		SrcIdFieldForError: "role_id",
		AssociatedIds:      change.add,
		DisassociatedIds:   change.remove,
		BeforeInsert:       newAssignmentIds(change.assignIdField),
	})
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	// One rebuild covers both directions: it drops every cached row this role's entitlements
	// produced and recomputes them, so removed principals lose them and added ones gain them.
	if err := this.permRepo.RebuildUserPermissionsForRole(ctx, change.roleId); err != nil {
		return nil, err
	}
	if err := this.auditRolePrincipals(ctx, change); err != nil {
		return nil, err
	}
	return result, nil
}

// assertRoleAssignable refuses a system role: it is held by every user of the tenant without
// an assignment row, so an assignment would be meaningless and a removal impossible.
func (this *RoleDomainServiceImpl) assertRoleAssignable(
	ctx corectx.Context, roleId model.Id,
) (ft.ClientErrors, error) {
	roleRes, err := corecrud.GetOne[domain.Role](ctx, corecrud.GetOneParam{
		Action:       "check role is assignable",
		DbRepoGetter: this.roleRepo,
		Query:        dyn.GetOneQuery{Id: roleId, Fields: []string{domain.RoleFieldIsSystem}},
	})
	if err != nil {
		return nil, err
	}
	if roleRes.ClientErrors.Count() > 0 {
		roleRes.ClientErrors.RenameField(basemodel.FieldId, "role_id")
		return roleRes.ClientErrors, nil
	}
	if !roleRes.HasData {
		return ft.ClientErrors{*ft.NewNotFoundError("role_id")}, nil
	}
	isSystem := roleRes.Data.GetFieldData().GetBool(domain.RoleFieldIsSystem)
	if isSystem != nil && *isSystem {
		return ft.ClientErrors{*ft.NewBusinessViolation(
			"role_id", ft.ErrorKey("err_system_role_assignment_not_allowed", "iam"),
			"a system role is held by every user and cannot be assigned or removed",
		)}, nil
	}
	return nil, nil
}

// auditRolePrincipals records one row per principal, as the principal-side endpoints do.
// Group rows name the group, not its members, for the same reason they do there.
func (this *RoleDomainServiceImpl) auditRolePrincipals(ctx corectx.Context, change rolePrincipalChange) error {
	err := this.auditor.recordRoleTransition(
		ctx, domain.PermissionHistoryEffectGrant, change.addReason, change.roleId, change.add.ToSlice(),
	)
	if err != nil {
		return err
	}
	return this.auditor.recordRoleTransition(
		ctx, domain.PermissionHistoryEffectRevoke, change.removeReason, change.roleId, change.remove.ToSlice(),
	)
}

func newAssignmentIds(idField string) func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
	return func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
		for _, record := range dbRecords {
			assignmentId, err := model.NewId()
			if err != nil {
				return err
			}
			record[idField] = *assignmentId
		}
		return nil
	}
}
