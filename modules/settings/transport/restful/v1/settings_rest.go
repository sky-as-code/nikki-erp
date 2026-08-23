package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

type settingsRestParams struct {
	dig.In

	TenantSettingsSvc it.TenantSettingsAppService
	OrgSettingsSvc    it.OrgSettingsAppService
	UserPrefsSvc      it.UserPreferencesAppService
}

func NewSettingsRest(params settingsRestParams) *SettingsRest {
	return &SettingsRest{
		tenantSettingsSvc: params.TenantSettingsSvc,
		orgSettingsSvc:    params.OrgSettingsSvc,
		userPrefsSvc:      params.UserPrefsSvc,
	}
}

// SettingsRest exposes the read and update operations of the three per-level application services.
//
// There is one pair of endpoints per level rather than one pair taking a level parameter, mirroring
// the three separate service contracts: the level a caller acts at is decided by the route they
// reached, never by a value they send, so a request body cannot widen the caller's reach.
//
// Authorization is not performed here. Each application service asserts the caller's permission for
// its own level, which is the only place that knows what the level means.
type SettingsRest struct {
	httpserver.RestBase

	tenantSettingsSvc it.TenantSettingsAppService
	orgSettingsSvc    it.OrgSettingsAppService
	userPrefsSvc      it.UserPreferencesAppService
}

func (this SettingsRest) GetTenantSettings(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get tenant settings"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tenantSettingsSvc.GetTenantSettings,
		GetSettingsRequest.ToQuery,
		NewGetSettingsResponse,
		httpserver.JsonOk,
	)
}

func (this SettingsRest) SetTenantSettings(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set tenant settings"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.tenantSettingsSvc.SetTenantSettings,
		SetSettingsRequest.ToCommand,
		NewSetSettingsResponse,
		httpserver.JsonOk,
	)
}

func (this SettingsRest) GetOrgSettings(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get org settings"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.orgSettingsSvc.GetOrgSettings,
		GetSettingsRequest.ToQuery,
		NewGetSettingsResponse,
		httpserver.JsonOk,
	)
}

func (this SettingsRest) SetOrgSettings(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set org settings"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.orgSettingsSvc.SetOrgSettings,
		SetSettingsRequest.ToCommand,
		NewSetSettingsResponse,
		httpserver.JsonOk,
	)
}

func (this SettingsRest) GetUserPreferences(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get user preferences"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.userPrefsSvc.GetUserPreferences,
		GetSettingsRequest.ToQuery,
		NewGetSettingsResponse,
		httpserver.JsonOk,
	)
}

func (this SettingsRest) SetUserPreferences(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set user preferences"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.userPrefsSvc.SetUserPreferences,
		SetSettingsRequest.ToCommand,
		NewSetSettingsResponse,
		httpserver.JsonOk,
	)
}
