package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

func (this UserRest) CreateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create user",
		echoCtx,
		&it.CreateUserCommand{},
		this.UserSvc.CreateUser,
	)
}

func (this UserRest) DeleteUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete user",
		echoCtx,
		this.UserSvc.DeleteUser,
	)
}

func (this UserRest) GetUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get user",
		echoCtx,
		this.UserSvc.GetUser,
	)
}

func (this UserRest) SearchUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search users",
		echoCtx,
		this.UserSvc.SearchUsers,
	)
}

func (this UserRest) SetUserIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set user is_archived",
		echoCtx,
		this.UserSvc.SetUserIsArchived,
	)
}

func (this UserRest) UpdateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update user",
		echoCtx,
		&it.UpdateUserCommand{},
		this.UserSvc.UpdateUser,
	)
}

func (this UserRest) UserExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"user exists",
		echoCtx,
		this.UserSvc.UserExists,
	)
}

/*
 * Non-CRUD APIs
 */

func (this UserRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.UserSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}

// func (this UserRest) GetUserContextUser(echoCtx *echo.Context) (err error) {
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
