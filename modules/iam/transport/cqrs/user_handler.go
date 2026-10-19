package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dmodel "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

func NewUserHandler(userSvc it.UserCrudDomainService) *UserHandler {
	return &UserHandler{
		UserSvc: userSvc,
	}
}

type UserHandler struct {
	UserSvc it.UserCrudDomainService
}

func (this *UserHandler) GetEnabledUser(
	ctx context.Context, packet *cqrs.RequestPacket[it.GetUserQuery],
) (*cqrs.Reply[dmodel.OpResult[models.User]], error) {
	return cqrs.ServePacket(ctx, string(c.IamModuleName), packet, this.UserSvc.GetEnabledUser)
}

