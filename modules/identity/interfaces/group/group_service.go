package group

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type GroupDomainService interface {
	CreateGroup(ctx corectx.Context, cmd CreateGroupCommand, opts ...corecrud.ServiceCreateOptions[*domain.Group]) (*CreateGroupResult, error)
	DeleteGroup(ctx corectx.Context, cmd DeleteGroupCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteGroupResult, error)
	GetGroup(ctx corectx.Context, query GetGroupQuery) (*dyn.OpResult[domain.Group], error)
	GroupExists(ctx corectx.Context, query GroupExistsQuery) (*GroupExistsResult, error)
	ManageGroupUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	SearchGroups(ctx corectx.Context, query SearchGroupsQuery, opts ...corecrud.ServiceSearchOptions) (*SearchGroupsResult, error)
	UpdateGroup(ctx corectx.Context, cmd UpdateGroupCommand, opts ...corecrud.ServiceUpdateOptions[*domain.Group]) (*UpdateGroupResult, error)
}

type GroupAppService interface {
	CreateGroup(ctx corectx.Context, cmd CreateGroupCommand) (*CreateGroupResult, error)
	DeleteGroup(ctx corectx.Context, cmd DeleteGroupCommand) (*DeleteGroupResult, error)
	GetGroup(ctx corectx.Context, query GetGroupQuery) (*GetGroupResult, error)
	GroupExists(ctx corectx.Context, query GroupExistsQuery) (*GroupExistsResult, error)
	ManageGroupUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	SearchGroups(ctx corectx.Context, query SearchGroupsQuery) (*SearchGroupsResult, error)
	UpdateGroup(ctx corectx.Context, cmd UpdateGroupCommand) (*UpdateGroupResult, error)
}
