package group

import (
	"context"
)

type GroupService interface {
	CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
	DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) (*DeleteGroupResult, error)
	GetGroupById(ctx context.Context, cmd GetGroupByIdQuery) (*GetGroupByIdResult, error)
}
