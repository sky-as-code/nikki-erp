package user

import (
    "context"

    "github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type CommandHandler struct {
    repo      Repository
    eventBus  *cqrs.EventBus
}

func NewCommandHandler(repo Repository, eventBus *cqrs.EventBus) *CommandHandler {
    return &CommandHandler{
        repo:     repo,
        eventBus: eventBus,
    }
}

func (thisHandler *CommandHandler) HandleCreateUser(ctx context.Context, cmd *CreateUserCommand) error {
    if err := thisHandler.repo.Create(ctx, cmd); err != nil {
        return err
    }

    event := &UserCreatedEvent{
        ID:          cmd.ID,
        Username:    cmd.Username,
        Email:       cmd.Email,
        DisplayName: cmd.DisplayName,
        AvatarURL:   cmd.AvatarURL,
        Status:      cmd.Status,
        CreatedBy:   cmd.CreatedBy,
        EventID:     NewEventID(),
    }

    return thisHandler.eventBus.Publish(ctx, event)
}

func (thisHandler *CommandHandler) HandleUpdateUser(ctx context.Context, cmd *UpdateUserCommand) error {
    if err := thisHandler.repo.Update(ctx, cmd); err != nil {
        return err
    }

    event := &UserUpdatedEvent{
        ID:          cmd.ID,
        DisplayName: cmd.DisplayName,
        AvatarURL:   cmd.AvatarURL,
        Status:      cmd.Status,
        UpdatedBy:   cmd.UpdatedBy,
        EventID:     NewEventID(),
    }

    return thisHandler.eventBus.Publish(ctx, event)
}

func (thisHandler *CommandHandler) HandleDeleteUser(ctx context.Context, cmd *DeleteUserCommand) error {
    if err := thisHandler.repo.Delete(ctx, cmd.ID, cmd.DeletedBy); err != nil {
        return err
    }

    event := &UserDeletedEvent{
        ID:        cmd.ID,
        DeletedBy: cmd.DeletedBy,
        EventID:   NewEventID(),
    }

    return thisHandler.eventBus.Publish(ctx, event)
}

func (thisHandler *CommandHandler) HandleGetUserByID(ctx context.Context, query *GetUserByIDQuery) (*User, error) {
    return thisHandler.repo.FindByID(ctx, query.ID)
}

func (thisHandler *CommandHandler) HandleGetUserByUsername(ctx context.Context, query *GetUserByUsernameQuery) (*User, error) {
    return thisHandler.repo.FindByUsername(ctx, query.Username)
}

func (thisHandler *CommandHandler) HandleGetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*User, error) {
    return thisHandler.repo.FindByEmail(ctx, query.Email)
}