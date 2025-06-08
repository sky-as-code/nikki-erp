package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserHandler(userSvc it.UserService, logger logging.LoggerService) *UserHandler {
	return &UserHandler{
		Logger:  logger,
		UserSvc: userSvc,
	}
}

type UserHandler struct {
	Logger  logging.LoggerService
	UserSvc it.UserService
}

func (this *UserHandler) Create(ctx context.Context, packet *cqrs.RequestPacket[it.CreateUserCommand]) (*cqrs.Reply[it.CreateUserResult], error) {
	cmd := packet.Request()
	result, err := this.UserSvc.CreateUser(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateUserResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *UserHandler) Update(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUserCommand]) (*cqrs.Reply[it.UpdateUserResult], error) {
	cmd := packet.Request()
	result, err := this.UserSvc.UpdateUser(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateUserResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *UserHandler) Delete(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUserCommand]) (*cqrs.Reply[it.DeleteUserResult], error) {
	cmd := packet.Request()
	result, err := this.UserSvc.DeleteUser(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[it.DeleteUserResult]{
		Result: *result,
	}, nil
}

func (this *UserHandler) GetUserById(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserByIdQuery]) (*cqrs.Reply[it.GetUserByIdResult], error) {
	cmd := packet.Request()
	result, err := this.UserSvc.GetUserById(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetUserByIdResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *UserHandler) SearchUsers(ctx context.Context, packet *cqrs.RequestPacket[it.SearchUsersCommand]) (*cqrs.Reply[it.SearchUsersResult], error) {
	cmd := packet.Request()
	result, err := this.UserSvc.SearchUsers(ctx, *cmd)
	if err != nil {
		return nil, err
	}

	reply := &cqrs.Reply[it.SearchUsersResult]{
		Result: *result,
	}
	return reply, nil
}

// func (this *UserCommandHandler) HandleGetUserByUsername(ctx context.Context, query *GetUserByUsernameQuery) (*User, error) {
// 	return this.repo.FindByUsername(ctx, query.Username)
// }

// func (this *UserCommandHandler) HandleGetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*User, error) {
// 	return this.repo.FindByEmail(ctx, query.Email)
// }
