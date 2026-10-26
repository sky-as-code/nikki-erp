package settings

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// SettingsAppService is Sales's own update of the settings it declares, served at
// `PATCH /v1/sales/settings/:level`.
type SettingsAppService interface {
	UpdateSettings(
		ctx corectx.Context, cmd coreSettings.UpdateSettingsCommand,
	) (*coreSettings.SetSettingsResult, error)
}
