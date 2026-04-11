package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserHandler(userSvc it.UserService) *UserHandler {
	return &UserHandler{
		UserSvc: userSvc,
	}
}

type UserHandler struct {
	UserSvc it.UserService
}

func (this *UserHandler) CreateUser(ctx context.Context, packet *cqrs.RequestPacket[it.CreateUserCommand]) (*cqrs.Reply[it.CreateUserResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.CreateUser)
}

func (this *UserHandler) UpdateUser(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUserCommand]) (*cqrs.Reply[it.UpdateUserResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.UpdateUser)
}

func (this *UserHandler) DeleteUser(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUserCommand]) (*cqrs.Reply[it.DeleteUserResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.DeleteUser)
}

func (this *UserHandler) GetUser(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserQuery]) (*cqrs.Reply[it.GetUserResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.GetUser)
}

func (this *UserHandler) GetEnabledUser(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserQuery]) (*cqrs.Reply[it.GetUserResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.GetEnabledUser)
}

func (this *UserHandler) SearchUsers(ctx context.Context, packet *cqrs.RequestPacket[it.SearchUsersQuery]) (*cqrs.Reply[it.SearchUsersResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.SearchUsers)
}

func (this *UserHandler) UserExists(ctx context.Context, packet *cqrs.RequestPacket[it.UserExistsQuery]) (*cqrs.Reply[it.UserExistsResult], error) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.UserSvc.UserExists)
}
