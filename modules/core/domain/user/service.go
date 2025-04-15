package user

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Service struct {
	commandBus *cqrs.CommandBus
}

func NewService(commandBus *cqrs.CommandBus) *Service {
	return &Service{
		commandBus: commandBus,
	}
}

func (thisSvc *Service) CreateUser(ctx context.Context, cmd *CreateUserCommand) error {
	return thisSvc.commandBus.Send(ctx, cmd)
}

func (thisSvc *Service) UpdateUser(ctx context.Context, cmd *UpdateUserCommand) error {
	return thisSvc.commandBus.Send(ctx, cmd)
}

func (thisSvc *Service) DeleteUser(ctx context.Context, id string, deletedBy string) error {
	cmd := &DeleteUserCommand{
		ID:        id,
		DeletedBy: deletedBy,
	}
	return thisSvc.commandBus.Send(ctx, cmd)
}

func (thisSvc *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := &GetUserByIDQuery{ID: id}
	user, err := thisSvc.commandBus.Send(ctx, query)
	if err != nil {
		return nil, err
	}
	return user.(*User), nil
}

func (thisSvc *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := &GetUserByUsernameQuery{Username: username}
	user, err := thisSvc.commandBus.Send(ctx, query)
	if err != nil {
		return nil, err
	}
	return user.(*User), nil
}

func (thisSvc *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := &GetUserByEmailQuery{Email: email}
	user, err := thisSvc.commandBus.Send(ctx, query)
	if err != nil {
		return nil, err
	}
	return user.(*User), nil
}
