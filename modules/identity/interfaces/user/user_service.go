package user

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

type UserService interface {
	CreateUser(ctx crud.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	CreateUser2(ctx dEnt.Context, cmd CreateUserCommand2) (*CreateUserResult2, error)
	DeleteUser(ctx crud.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	Exists(ctx crud.Context, cmd UserExistsQuery) (*UserExistsResult, error)
	ExistsMulti(ctx crud.Context, cmd UserExistsMultiQuery) (*UserExistsMultiResult, error)
	GetUserById(ctx crud.Context, query GetUserByIdQuery) (*GetUserByIdResult, error)
	GetUserByEmail(ctx crud.Context, query GetUserByEmailQuery) (*GetUserByEmailResult, error)
	MustGetActiveUser(ctx crud.Context, query MustGetActiveUserQuery) (*MustGetActiveUserResult, error)
	SearchUsers(ctx crud.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	UpdateUser(ctx crud.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
	// FindDirectApprover(ctx crud.Context, query FindDirectApproverQuery) (*FindDirectApproverResult, error)
	GetUserContext(ctx crud.Context, query GetUserContextQuery) (*GetUserContextResultData, error)
}
