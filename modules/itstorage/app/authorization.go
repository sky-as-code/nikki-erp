package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// assertFileRead authorizes reading a stored file by the org its record carries. The record's
// own org is the scope, so there is no caller-named org to check membership of: a caller without
// a read grant in that org is refused as unauthorized.
func assertFileRead(ctx corectx.Context, svc composable.CrudApplicationService, orgId *model.Id) *ft.ClientErrors {
	if orgId != nil {
		return svc.AssertPermission(ctx, composable.PermissionRead, orgId)
	}
	return assertFileAction(ctx, svc, composable.PermissionRead, nil)
}

// assertFileAction authorizes actionCode on a file of orgId, or on a tenant-wide file when orgId
// is nil.
//
// An org file goes through AssertAction, which also requires the caller to belong to the org. A
// tenant-wide file belongs to no org, so it is granted to a caller holding the action tenant-wide
// or in any org they belong to: a user's avatar must be readable by colleagues whose grants are
// all org-scoped.
func assertFileAction(
	ctx corectx.Context, svc composable.CrudApplicationService, actionCode string, orgId *model.Id,
) *ft.ClientErrors {
	if orgId != nil {
		_, cErrs := svc.AssertAction(ctx, actionCode, dmodel.DynamicFields{basemodel.FieldOrgId: *orgId})
		return cErrs
	}
	refused := svc.AssertPermission(ctx, actionCode, nil)
	if refused == nil {
		return nil
	}
	perms := ctx.GetPermissions()
	if perms.Principal.Kind == corectx.PrincipalKindService && perms.Principal.OrgId != nil {
		return svc.AssertPermission(ctx, actionCode, perms.Principal.OrgId)
	}
	if perms.UserOrgIds == nil {
		return refused
	}
	for _, userOrgId := range perms.UserOrgIds.ToSlice() {
		if svc.AssertPermission(ctx, actionCode, &userOrgId) == nil {
			return nil
		}
	}
	return refused
}
