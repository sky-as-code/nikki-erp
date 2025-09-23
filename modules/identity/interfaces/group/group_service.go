package group

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GroupService interface {
	AddRemoveUsers(ctx crud.Context, cmd AddRemoveUsersCommand) (*AddRemoveUsersResult, error)
	CreateGroup(ctx crud.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	DeleteGroup(ctx crud.Context, cmd DeleteGroupCommand) (*DeleteGroupResult, error)
	GetGroupById(ctx crud.Context, query GetGroupByIdQuery) (*GetGroupByIdResult, error)
	SearchGroups(ctx crud.Context, query SearchGroupsQuery) (*SearchGroupsResult, error)
	UpdateGroup(ctx crud.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
	Exist(ctx crud.Context, cmd GroupExistsCommand) (*GroupExistsResult, error)
}
