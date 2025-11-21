package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type userRestParams struct {
	dig.In

	UserSvc it.UserService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		UserSvc: params.UserSvc,
	}
}

type UserRest struct {
	httpserver.RestBase
	UserSvc it.UserService
}

func init() {
	schema.AdhocRegistry().Add(
		"identity.createUserRequest",
		schema.DefineAdhoc().
			Field(schema.CloneField("identity.user", "display_name").Required()).
			FieldHolder("contact", true,
				schema.DefineAdhoc().
					Field(schema.CloneField("identity.user", "email").Required()).
					Field(schema.CloneField("identity.user", "hierarchy_id").Required()),
			).
			Build(),
	)
}

func (this UserRest) DynamicCreateUser(request any, echoCtx echo.Context) (err error) {
	util.Unused(request)
	return nil
}

func (this UserRest) CreateUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create user"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.CreateUser,
		func(request CreateUserRequest) it.CreateUserCommand {
			return it.CreateUserCommand(request)
		},
		func(result it.CreateUserResult) CreateUserResponse {
			response := CreateUserResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this UserRest) UpdateUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update user"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.UpdateUser,
		func(request UpdateUserRequest) it.UpdateUserCommand {
			return it.UpdateUserCommand(request)
		},
		func(result it.UpdateUserResult) UpdateUserResponse {
			response := UpdateUserResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) DeleteUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete user"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.DeleteUser,
		func(request DeleteUserRequest) it.DeleteUserCommand {
			return it.DeleteUserCommand(request)
		},
		func(result it.DeleteUserResult) DeleteUserResponse {
			response := DeleteUserResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) GetUserById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.GetUserById,
		func(request GetUserByIdRequest) it.GetUserByIdQuery {
			return it.GetUserByIdQuery(request)
		},
		func(result it.GetUserByIdResult) GetUserByIdResponse {
			response := GetUserByIdResponse{}
			response.FromUser(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) SearchUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search users"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.SearchUsers,
		func(request SearchUsersRequest) it.SearchUsersQuery {
			return it.SearchUsersQuery(request)
		},
		func(result it.SearchUsersResult) SearchUsersResponse {
			response := SearchUsersResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) UserExistsMulti(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST user exists multi"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.ExistsMulti,
		func(request UserExistsMultiRequest) it.UserExistsMultiQuery {
			return it.UserExistsMultiQuery(request)
		},
		func(result it.UserExistsMultiResult) UserExistsMultiResponse {
			return *result.Data
		},
		httpserver.JsonOk,
	)
	return err
}
