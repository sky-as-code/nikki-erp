package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

type userPreferenceRestParams struct {
	dig.In

	UserPreferenceSvc it.UserPreferenceApplicationService
}

func NewUserPreferenceRest(params userPreferenceRestParams) *UserPreferenceRest {
	return &UserPreferenceRest{
		UserPreferenceSvc: params.UserPreferenceSvc,
	}
}

type UserPreferenceRest struct {
	httpserver.RestBase
	UserPreferenceSvc it.UserPreferenceApplicationService
}

func (this UserPreferenceRest) CreateUserPreference(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create user preference"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.CreateUserPreference,
		func(request CreateUserPreferenceRequest) it.CreateUserPreferenceCommand {
			return it.CreateUserPreferenceCommand(request)
		},
		func(data domain.UserPreference) CreateUserPreferenceResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this UserPreferenceRest) DeleteUserPreference(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete user preference"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.DeleteUserPreference,
		func(request DeleteUserPreferenceRequest) it.DeleteUserPreferenceCommand {
			return it.DeleteUserPreferenceCommand(request)
		},
		func(data dyn.MutateResultData) DeleteUserPreferenceResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this UserPreferenceRest) GetUserPreference(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user preference"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.GetUserPreference,
		func(request GetUserPreferenceRequest) it.GetUserPreferenceQuery {
			return it.GetUserPreferenceQuery(request)
		},
		func(data domain.UserPreference) GetUserPreferenceResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this UserPreferenceRest) UserPreferenceExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST user preference exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.UserPreferenceExists,
		func(request UserPreferenceExistsRequest) it.UserPreferenceExistsQuery {
			return it.UserPreferenceExistsQuery(request)
		},
		func(data dyn.ExistsResultData) UserPreferenceExistsResponse {
			return UserPreferenceExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this UserPreferenceRest) SearchUserPreferences(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search user preferences"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.SearchUserPreferences,
		func(request SearchUserPreferencesRequest) it.SearchUserPreferencesQuery {
			return it.SearchUserPreferencesQuery(request)
		},
		func(data it.SearchUserPreferencesResultData) SearchUserPreferencesResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
	)
}

func (this UserPreferenceRest) UpdateUserPreference(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update user preference"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.UserPreferenceSvc.UpdateUserPreference,
		func(request UpdateUserPreferenceRequest) it.UpdateUserPreferenceCommand {
			return it.UpdateUserPreferenceCommand(request)
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this UserPreferenceRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.UserPreferenceSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
