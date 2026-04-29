package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type userRestParams struct {
	dig.In

	UserSvc it.UserAppService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		UserSvc: params.UserSvc,
	}
}

type UserRest struct {
	UserSvc it.UserAppService
}

func (this UserRest) CreateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateUserRequest, CreateUserResponse, domain.User](
		"create user",
		echoCtx,
		&it.CreateUserCommand{},
		this.UserSvc.CreateUser,
	)
}

func (this UserRest) DeleteUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteUserRequest, DeleteUserResponse](
		"delete user",
		echoCtx,
		this.UserSvc.DeleteUser,
	)
}

func (this UserRest) GetUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetUserRequest, GetUserResponse, domain.User](
		"get user",
		echoCtx,
		this.UserSvc.GetUser,
	)
}

func (this UserRest) SearchUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchUsersRequest, SearchUsersResponse, domain.User](
		"search users",
		echoCtx,
		this.UserSvc.SearchUsers,
	)
}

func (this UserRest) SetUserIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[SetUserIsArchivedRequest, SetUserIsArchivedResponse](
		"set user is_archived",
		echoCtx,
		this.UserSvc.SetUserIsArchived,
	)
}

func (this UserRest) UpdateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateUserRequest, UpdateUserResponse](
		"update user",
		echoCtx,
		&it.UpdateUserCommand{},
		this.UserSvc.UpdateUser,
	)
}

func (this UserRest) UserExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[UserExistsRequest, UserExistsResponse](
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

func (this UserRest) GetUserContext(echoCtx *echo.Context) (err error) {
	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	userPerm := reqCtx.GetPermissions()
	user := models.NewUserFrom(reqCtx.GetUser())
	echoCtx.JSON(http.StatusOK, GetUserContextResponse{
		Id:           string(user.MustGetId()),
		AvatarUrl:    user.MustGetAvatarUrl(),
		DisplayName:  user.MustGetDisplayName(),
		Email:        user.MustGetEmail(),
		Entitlements: userPerm.Entitlements.ToSlice(),
		OrgIds:       userPerm.UserOrgIds.ToSlice(),
	})
	return nil
}
