package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

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

func (this UserRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create user",
		echoCtx,
		&it.CreateUserCommand{},
		this.UserSvc.CreateUser,
	)
}

func (this UserRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete user",
		echoCtx,
		this.UserSvc.DeleteUser,
	)
}

func (this UserRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get user",
		echoCtx,
		this.UserSvc.GetUser,
	)
}

func (this UserRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search users",
		echoCtx,
		this.UserSvc.SearchUsers,
		true,
	)
}

func (this UserRest) SetIsArchived(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set user is_archived",
		echoCtx,
		this.UserSvc.SetUserIsArchived,
	)
}

func (this UserRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update user",
		echoCtx,
		&it.UpdateUserCommand{},
		this.UserSvc.UpdateUser,
	)
}

func (this UserRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"user exists",
		echoCtx,
		this.UserSvc.UserExists,
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
