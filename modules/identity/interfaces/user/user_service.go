package user

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UserService interface {
	DeleteUser(ctx crud.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	Exists(ctx crud.Context, cmd UserExistsQuery) (*UserExistsResult, error)
	ExistsMulti(ctx crud.Context, cmd UserExistsMultiQuery) (*UserExistsMultiResult, error)
	GetUserByEmail(ctx crud.Context, query GetUserByEmailQuery) (*GetUserByEmailResult, error)
	MustGetActiveUser(ctx crud.Context, query MustGetActiveUserQuery) (*MustGetActiveUserResult, error)
	SearchUsers(ctx crud.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	GetUserContext(ctx crud.Context, query GetUserContextQuery) (*GetUserContextResultData, error)

	ArchiveUser(ctx corectx.Context, cmd ArchiveUserCommand2) (*ArchiveUserResult2, error)
	CreateUser(ctx corectx.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	UpdateUser(ctx corectx.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
	GetOne(ctx corectx.Context, query GetUserQuery) (*GetUserResult, error)
	SearchUsers2(ctx corectx.Context, query SearchUsersQuery2) (*SearchUsersResult2, error)
}
