package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
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

func (this *UserHandler) Create(ctx context.Context, packet *cqrs.RequestPacket[it.CreateUserCommand]) (*cqrs.Reply[it.CreateUserResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.CreateUser)
}

func (this *UserHandler) Update(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUserCommand]) (*cqrs.Reply[it.UpdateUserResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.UpdateUser)
}

func (this *UserHandler) Delete(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUserCommand]) (*cqrs.Reply[it.DeleteUserResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.DeleteUser)
}

func (this *UserHandler) GetUserById(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserByIdQuery]) (*cqrs.Reply[it.GetUserByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.GetUserById)
}

func (this *UserHandler) GetUserByEmail(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserByEmailQuery]) (*cqrs.Reply[it.GetUserByEmailResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.GetUserByEmail)
}

func (this *UserHandler) MustGetActiveUser(ctx context.Context, packet *cqrs.RequestPacket[it.MustGetActiveUserQuery]) (*cqrs.Reply[it.MustGetActiveUserResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.MustGetActiveUser)
}

func (this *UserHandler) SearchUsers(ctx context.Context, packet *cqrs.RequestPacket[it.SearchUsersQuery]) (*cqrs.Reply[it.SearchUsersResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.SearchUsers)
}

func (this *UserHandler) UserExists(ctx context.Context, packet *cqrs.RequestPacket[it.UserExistsCommand]) (*cqrs.Reply[it.UserExistsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.Exists)
}

func (this *UserHandler) UserExistsMulti(ctx context.Context, packet *cqrs.RequestPacket[it.UserExistsMultiCommand]) (*cqrs.Reply[it.UserExistsMultiResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.UserSvc.ExistsMulti)
}
