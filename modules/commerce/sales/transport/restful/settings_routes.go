package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	itSettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/settings"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

func initSettingsV1() error {
	return deps.Invoke(func(route *echo.Group, settingsSvc itSettings.SettingsAppService) {
		route.Group("/v1/"+modconstants.SettingsModuleKey).PATCH(coreSettings.UpdateSettingsRoutePath,
			coreSettings.UpdateSettingsHandler(settingsSvc.UpdateSettings), m.UserSmokeAuthz())
	})
}
