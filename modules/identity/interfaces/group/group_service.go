package group

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type GroupService interface {
	CreateGroup(ctx corectx.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	DeleteGroup(ctx corectx.Context, cmd DeleteGroupCommand) (*DeleteGroupResult, error)
	GetGroup(ctx corectx.Context, query GetGroupQuery) (*GetGroupResult, error)
	GroupExists(ctx corectx.Context, query GroupExistsQuery) (*GroupExistsResult, error)
	ManageGroupUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	SearchGroups(ctx corectx.Context, query SearchGroupsQuery) (*SearchGroupsResult, error)
	UpdateGroup(ctx corectx.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
}
