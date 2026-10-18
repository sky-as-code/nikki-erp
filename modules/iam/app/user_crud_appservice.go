package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/filefield"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

const userAvatarKeyPrefix = "iam/users"

// UserFileFields lists the uploadable files of a user.
func UserFileFields() []filefield.FileField {
	return []filefield.FileField{
		filefield.ImageField(models.UserFileAvatar, models.UserFieldAvatarFileKey,
			models.UserFieldAvatarFileUrl, userAvatarKeyPrefix),
	}
}

// NewUserCrudApplicationService is handed the composable default by the iam_user onion.
func NewUserCrudApplicationService(
	base composable.CrudApplicationService, storage filestorage.FileStorageAdapter,
) itUser.UserCrudApplicationService {
	return &UserCrudApplicationServiceImpl{
		ApplicationService: filefield.NewApplicationService(base, storage, UserFileFields()),
	}
}

// UserCrudApplicationServiceImpl is the authorized CRUD of the resource, with the avatar file
// field presigned on read and written only through files/avatar.
type UserCrudApplicationServiceImpl struct {
	filefield.ApplicationService
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
