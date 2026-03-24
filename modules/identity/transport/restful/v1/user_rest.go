package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	middleWare "github.com/sky-as-code/nikki-erp/common/middleware"
	"github.com/sky-as-code/nikki-erp/common/modelmapper"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
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

func (this UserRest) CreateUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create user"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequestDynamic(
		echoCtx,
		this.UserSvc.CreateUser2,
		func(requestFields schema.DynamicFields) it.CreateUserCommand2 {
			cmd := it.CreateUserCommand2{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.UserEntity) CreateUserResponse {
			response, err := modelmapper.MapToStruct[*CreateUserResponse](data.GetFieldData())
			ft.PanicOnErr(err)
			return *response
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
		func(request GetUserByIdRequest) it.GetUser {
			return it.GetUser(request)
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

func (this UserRest) UpdateUser2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update user 2"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.UserSvc.UpdateUser2,
		func(requestFields schema.DynamicFields) it.UpdateUserCommand2 {
			cmd := it.UpdateUserCommand2{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.UserEntity) UpdateUser2Response {
			response, err := modelmapper.MapToStruct[*UpdateUser2Response](data.GetFieldData())
			ft.PanicOnErr(err)
			return *response
		},
		httpserver.JsonOk,
	)
}

func (this UserRest) GetOne(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user by ID"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.GetOne,
		func(request it.GetUser) it.GetUser {
			return request
		},
		func(data domain.UserEntity) schema.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this UserRest) ArchiveUser2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST archive user 2"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.UserSvc.ArchiveUser2,
		func(requestFields schema.DynamicFields) it.ArchiveUserCommand2 {
			cmd := it.ArchiveUserCommand2{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.UserEntity) ArchiveUser2Response {
			response, err := modelmapper.MapToStruct[*ArchiveUser2Response](data.GetFieldData())
			ft.PanicOnErr(err)
			return *response
		},
		httpserver.JsonOk,
	)
}

func (this UserRest) SearchUsers2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search users 2"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.SearchUsers2,
		func(request SearchUsers2Request) it.SearchUsersQuery2 {
			return it.SearchUsersQuery2(request)
		},
		func(result dEnt.PagedResult[domain.UserEntity]) SearchUsers2Response {
			items := array.Map(result.Items, func(item domain.UserEntity) schema.DynamicFields {
				return item.GetFieldData()
			})
			return SearchUsers2Response{Items: items, Total: result.Total, Page: result.Page, Size: result.Size}
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) GetUserContext(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user context"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UserSvc.GetUserContext,
		func(request GetUserContextRequest) it.GetUserContextQuery {
			request.UserId = middleWare.GetUserIdFromContext(echoCtx.Request().Context())
			return it.GetUserContextQuery(request)
		},
		func(result it.GetUserContextResultData) GetUserContextResponse {
			return *result.Data
		},
		httpserver.JsonOk,
	)
	return err
}
