package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	middleWare "github.com/sky-as-code/nikki-erp/common/middleware"
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

func (this UserRest) CreateUser(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create user"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequestDynamic[CreateUserResponse](
		echoCtx,
		"create user",
		func() schema.DynamicModelSetter {
			return &CreateUserRequest{}
		},
		this.UserSvc.CreateUser2,
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

func (this UserRest) UpdateUser2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update user 2"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequestDynamic[UpdateUser2Response](
		echoCtx,
		"update user 2",
		func() schema.DynamicModelSetter { return &UpdateUser2Request{} },
		this.UserSvc.UpdateUser2,
		httpserver.JsonOk,
	)
}

func (this UserRest) GetUserByPk2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user by pk"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequestDynamic[GetUserByPk2Response](
		echoCtx,
		"get user by pk",
		func() schema.DynamicModelSetter { return &GetUserByPk2Request{} },
		this.UserSvc.GetUserByPk2,
		httpserver.JsonOk,
	)
}

func (this UserRest) ArchiveUser2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST archive user 2"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequestDynamic[ArchiveUser2Response](
		echoCtx,
		"archive user 2",
		func() schema.DynamicModelSetter { return &ArchiveUser2Request{} },
		this.UserSvc.ArchiveUser2,
		httpserver.JsonOk,
	)
}

func (this UserRest) SearchUsers2(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search users 2"); e != nil {
			err = e
		}
	}()
	// var query SearchUsers2Request
	// if err = echoCtx.Bind(&query); err != nil {
	// 	return err
	// }
	// reqCtx := echoCtx.Request().Context().(dynamicentity.Context)
	// result, err := this.UserSvc.SearchUsers2(reqCtx, query)
	// if err != nil {
	// 	return err
	// }
	// if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
	// 	return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	// }
	// return httpserver.JsonOk(echoCtx, toSearchUsers2Response(result.Data))
	return httpserver.JsonOk(echoCtx, nil)
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
