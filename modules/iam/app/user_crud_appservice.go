package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewUserCrudApplicationService is handed the composable default by the iam_user onion.
func NewUserCrudApplicationService(
	base composable.CrudApplicationService, domSvc itUser.UserCrudDomainService,
) itUser.UserCrudApplicationService {
	return &UserCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

// UserCrudApplicationServiceImpl is the authorized CRUD of the resource. The avatar is an
// ordinary avatar_file_path column; the domain service checks it against IT Storage.
type UserCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itUser.UserCrudDomainService
}

// ManageRoles asserts manage_role_assignments, the same permission the role side asserts for
// iam_role/:id/manage_users, so neither side opens a path the other keeps shut.
func (this *UserCrudApplicationServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itUser.ManageUserRolesCommand,
) (*itUser.ManageUserRolesResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd,
		composable.ActionPermission{ActionCode: actionManageRoleAssignments})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageRoles(ctx, cmd)
}

// SendInvitation answers with the address an invitation would go to. Sending the actual email is
// left to the invitation feature.
func (this *UserCrudApplicationServiceImpl) SendInvitation(
	ctx corectx.Context, cmd itUser.SendUserInvitationCommand,
) (*itUser.SendUserInvitationResult, error) {
	orgId, cErrs := this.AssertAction(ctx, composable.PermissionUpdate, cmd)
	if cErrs == nil && orgId != nil {
		var err error
		if cErrs, err = this.AssertRecordInOrg(ctx, cmd, *orgId); err != nil {
			return nil, err
		}
	}
	if cErrs != nil {
		return &itUser.SendUserInvitationResult{ClientErrors: *cErrs}, nil
	}

	vErrs := ft.ClientErrors{}
	found, err := this.FetchByKeys(ctx, dmodel.DynamicFields{models.UserFieldId: cmd[models.UserFieldId]}, &vErrs)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &itUser.SendUserInvitationResult{ClientErrors: vErrs}, nil
	}
	if found == nil {
		return &itUser.SendUserInvitationResult{}, nil
	}
	return &itUser.SendUserInvitationResult{
		Data:    dmodel.DynamicFields{models.UserFieldEmail: found[models.UserFieldEmail]},
		HasData: true,
	}, nil
}
