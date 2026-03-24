package user

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UserService interface {
	DeleteUser(ctx crud.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	Exists(ctx crud.Context, cmd UserExistsQuery) (*UserExistsResult, error)
	ExistsMulti(ctx crud.Context, cmd UserExistsMultiQuery) (*UserExistsMultiResult, error)
	GetUserById(ctx crud.Context, query GetUser) (*GetUserByIdResult, error)
	GetUserByEmail(ctx crud.Context, query GetUserByEmailQuery) (*GetUserByEmailResult, error)
	MustGetActiveUser(ctx crud.Context, query MustGetActiveUserQuery) (*MustGetActiveUserResult, error)
	SearchUsers(ctx crud.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	// UpdateUser(ctx crud.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
	// FindDirectApprover(ctx crud.Context, query FindDirectApproverQuery) (*FindDirectApproverResult, error)
	GetUserContext(ctx crud.Context, query GetUserContextQuery) (*GetUserContextResultData, error)

	ArchiveUser(ctx corectx.Context, cmd ArchiveUserCommand2) (*ArchiveUserResult2, error)
	CreateUser(ctx corectx.Context, cmd CreateUserCommand2) (*CreateUserResult2, error)
	UpdateUser(ctx corectx.Context, cmd UpdateUserCommand2) (*UpdateUserResult2, error)
	GetOne(ctx corectx.Context, query GetUser) (*GetUserResult, error)
	SearchUsers2(ctx corectx.Context, query SearchUsersQuery2) (*SearchUsersResult2, error)
}
