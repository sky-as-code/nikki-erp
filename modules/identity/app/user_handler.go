package app

import (
	"context"

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
	if err != nil {
		this.Logger.Error("failed to create user", err)
		return nil, err
	}
	reply := &cqrs.Reply[it.CreateUserResult]{
		Result: *result,
	}
	return reply, err

	// event := &UserCreatedEvent{
	// 	ID:          packet.Id,
	// 	Username:    packet.Username,
	// 	Email:       packet.Email,
	// 	DisplayName: packet.DisplayName,
	// 	AvatarURL:   packet.AvatarUrl,
	// 	Status:      packet.Status,
	// 	CreatedBy:   packet.CreatedBy,
	// 	EventID:     NewEventID(),
	// }

	// return this.eventBus.Publish(ctx, event)
}

func (this *UserHandler) Update(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateUserCommand]) error {
	return nil

	// event := &UserUpdatedEvent{
	// 	ID:          cmd.Id,
	// 	DisplayName: cmd.DisplayName,
	// 	AvatarURL:   cmd.AvatarUrl,
	// 	Status:      cmd.Status,
	// 	UpdatedBy:   cmd.UpdatedBy,
	// 	EventID:     NewEventID(),
	// }

	// return this.eventBus.Publish(ctx, event)
}

func (this *UserHandler) Delete(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteUserCommand]) error {
	return nil

	// event := &UserDeletedEvent{
	// 	ID:        cmd.Id,
	// 	DeletedBy: cmd.DeletedBy,
	// 	EventID:   NewEventID(),
	// }

	// return this.eventBus.Publish(ctx, event)
}

// func (this *UserCommandHandler) HandleGetUserByID(ctx context.Context, query *GetUserByIdQuery) (*User, error) {
// 	return this.repo.FindByID(ctx, query.Id)
// }

// func (this *UserCommandHandler) HandleGetUserByUsername(ctx context.Context, query *GetUserByUsernameQuery) (*User, error) {
// 	return this.repo.FindByUsername(ctx, query.Username)
// }

// func (this *UserCommandHandler) HandleGetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*User, error) {
// 	return this.repo.FindByEmail(ctx, query.Email)
// }
