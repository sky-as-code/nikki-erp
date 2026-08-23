package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/settings/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/settings/transport/restful/v1"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewSettingsRest),
		initSettingsV1(),
	)
}

// initSettingsV1 mounts the engine routes and the per-level settings routes under /v1/settings.
func initSettingsV1() error {
	return deps.Invoke(func(route *echo.Group, settingsRest *v1.SettingsRest) error {
		routeV1 := route.Group("/v1/settings")

		// The level routes are registered before the engines, because an engine mounts its CRUD
		// under the schema name and a later catch-all would otherwise shadow these static paths.
		registerLevelRoutes(routeV1, settingsRest)
		registerEngineRoutes(routeV1)
		return nil
	})
}

// registerLevelRoutes exposes read and update once per level. The level is part of the path rather
// than the payload, so that a caller holding org-level permission cannot reach the tenant level by
// sending a different value.
func registerLevelRoutes(routeV1 *echo.Group, settingsRest *v1.SettingsRest) {
	routeV1.GET("/tenant/:module_key", settingsRest.GetTenantSettings, m.SmokeAuthz())
	routeV1.PATCH("/tenant/:module_key", settingsRest.SetTenantSettings, m.SmokeAuthz())

	routeV1.GET("/org/:module_key", settingsRest.GetOrgSettings, m.SmokeAuthz())
	routeV1.PATCH("/org/:module_key", settingsRest.SetOrgSettings, m.SmokeAuthz())

	routeV1.GET("/user/:module_key", settingsRest.GetUserPreferences, m.SmokeAuthz())
	routeV1.PATCH("/user/:module_key", settingsRest.SetUserPreferences, m.SmokeAuthz())
}

// registerEngineRoutes exposes every Settings resource engine over HTTP.
// A missing engine is skipped, so that a build which drops one still starts.
func registerEngineRoutes(routeV1 *echo.Group) {
	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		engine, exists := dynamicresource.Registry().GetEngine(schemaName)
		if !exists {
			continue
		}
		engine.RestApi().RegisterRoutes(routeV1, m.SmokeAuthz())
	}
}
