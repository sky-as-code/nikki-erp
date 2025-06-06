package group

import (
	"context"
)

type GroupService interface {
	CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
	DeleteGroup(ctx context.Context, id string, deletedBy string) (*DeleteGroupResult, error)
	GetGroupByID(ctx context.Context, id string, withOrg bool) (*GetGroupByIdResult, error)
}
