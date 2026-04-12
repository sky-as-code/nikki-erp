package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	// middleWare "github.com/sky-as-code/nikki-erp/common/middleware"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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

func (this UserRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create user",
		echoCtx,
		&it.CreateUserCommand{},
		this.UserSvc.CreateUser,
	)
}

func (this UserRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeDelete[DeleteUserRequest](
		"delete user",
		echoCtx,
		this.UserSvc.DeleteUser,
	)
}

func (this UserRest) GetOne(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.GetUser,
		func(request GetUserRequest) it.GetUserQuery {
			return request
		},
		func(data domain.User) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this UserRest) Search(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search users 2"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.SearchUsers,
		func(request SearchUsers2Request) it.SearchUsersQuery {
			return it.SearchUsersQuery(request)
		},
		func(data it.SearchUsersResultData) SearchUsersResponse2 {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
	return err
}

func (this UserRest) SetIsArchived(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set user is_archived"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.SetUserIsArchived,
		func(request SetUserIsArchivedRequest) it.SetUserIsArchivedCommand {
			return request
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this UserRest) Update(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update user"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequestDynamic(
		echoCtx,
		this.UserSvc.UpdateUser,
		func(requestFields dmodel.DynamicFields) it.UpdateUserCommand {
			cmd := it.UpdateUserCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
	return err
}

func (this UserRest) Exists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST user exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserSvc.UserExists,
		func(request UserExistsRequest) it.UserExistsQuery {
			return it.UserExistsQuery(request)
		},
		func(data dyn.ExistsResultData) UserExistsResponse {
			return UserExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

// func (this UserRest) GetUserContext(echoCtx echo.Context) (err error) {
// 	defer func() {
// 		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user context"); e != nil {
// 			err = e
// 		}
// 	}()
// 	err = httpserver.ServeRequest(
// 		echoCtx, this.UserSvc.GetUserContext,
// 		func(request GetUserContextRequest) it.GetUserContextQuery {
// 			request.UserId = middleWare.GetUserIdFromContext(echoCtx.Request().Context())
// 			return it.GetUserContextQuery(request)
// 		},
// 		func(result it.GetUserContextResultData) GetUserContextResponse {
// 			return *result.Data
// 		},
// 		httpserver.JsonOk,
// 	)
// 	return err
// }
