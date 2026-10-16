package external

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

type SettingsServiceParams struct {
	dig.In

	TenantSvc itSettings.TenantSettingsAppService
	OrgSvc    itSettings.OrgSettingsAppService
	UserSvc   itSettings.UserPreferencesAppService
}

// NewSettingsServiceImpl points core's settings port at the Settings module's application
// services, one per level, so a write keeps the authorization each of them asserts.
//
// The Settings module lives in this binary, so the port is an in-process call. A build that
// split settings into its own service would register an HTTP client here instead.
func NewSettingsServiceImpl(params SettingsServiceParams) coreSettings.SettingsService {
	return &settingsServiceImpl{tenantSvc: params.TenantSvc, orgSvc: params.OrgSvc, userSvc: params.UserSvc}
}

type settingsServiceImpl struct {
	tenantSvc itSettings.TenantSettingsAppService
	orgSvc    itSettings.OrgSettingsAppService
	userSvc   itSettings.UserPreferencesAppService
}

func (this *settingsServiceImpl) SetSettings(
	ctx corectx.Context, level string, cmd coreSettings.SetSettingsCommand,
) (*coreSettings.SetSettingsResult, error) {
	switch level {
	case coreSettings.LevelTenant:
		return this.tenantSvc.SetTenantSettings(ctx, cmd)
	case coreSettings.LevelOrg:
		return this.orgSvc.SetOrgSettings(ctx, cmd)
	case coreSettings.LevelUser:
		return this.userSvc.SetUserPreferences(ctx, cmd)
	}
	return nil, errors.Errorf("SetSettings: unknown level '%s'", level)
}
