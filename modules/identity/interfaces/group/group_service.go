package group

import (
	"context"
)

type GroupService interface {
	AddRemoveUsers(ctx context.Context, cmd AddRemoveUsersCommand) (*AddRemoveUsersResult, error)
	CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) (*DeleteGroupResult, error)
	GetGroupById(ctx context.Context, query GetGroupByIdQuery) (*GetGroupByIdResult, error)
	SearchGroups(ctx context.Context, query SearchGroupsQuery) (*SearchGroupsResult, error)
	UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
	Exist(ctx context.Context, cmd GroupExistsCommand) (*GroupExistsResult, error)
}
