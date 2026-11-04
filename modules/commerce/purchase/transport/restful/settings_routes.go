package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itSettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/settings"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// initSettingsV1 serves PATCH /v1/purchase/settings/:level, Purchase's write of its own settings.
// Reads go through the Settings module (GET /v1/settings/org/purchase).
func initSettingsV1() error {
	return deps.Invoke(func(route *echo.Group, settingsSvc itSettings.SettingsAppService) {
		route.Group(routeV1Prefix).PATCH(coreSettings.UpdateSettingsRoutePath,
			coreSettings.UpdateSettingsHandler(settingsSvc.UpdateSettings), m.UserSmokeAuthz())
	})
}
