package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/settings"
)

func initSettingsV1() error {
	return deps.Invoke(func(route *echo.Group, settingsSvc itSettings.SettingsAppService) {
		route.Group("/v1/iam").PATCH(coreSettings.UpdateSettingsRoutePath,
			coreSettings.UpdateSettingsHandler(settingsSvc.UpdateSettings), m.SmokeAuthz())
	})
}
