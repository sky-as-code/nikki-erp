package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/settings/transport/restful/v1"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewSettingsRest, v1.NewSettingsSchemaRest, v1.NewSettingsRecordRest, v1.NewSettingsQuotaRest),
		initSettingsV1(),
	)
}

// initSettingsV1 mounts the per-level settings routes and the two read-only resources under
// /v1/settings.
func initSettingsV1() error {
	return deps.Invoke(func(
		route *echo.Group, settingsRest *v1.SettingsRest,
		schemaRest *v1.SettingsSchemaRest, recordRest *v1.SettingsRecordRest, quotaRest *v1.SettingsQuotaRest,
	) error {
		routeV1 := route.Group("/v1/settings")

		// The level routes are registered before the resources, because a resource mounts its
		// routes under the schema name and a later catch-all would otherwise shadow these paths.
		registerLevelRoutes(routeV1, settingsRest)
		return stdErr.Join(
			registerReadOnly(routeV1, models.SettingsSchemaSchemaName, schemaRest),
			registerReadOnly(routeV1, models.SettingsRecordSchemaName, recordRest),
			registerReadOnly(routeV1, models.SettingsQuotaSchemaName, quotaRest),
		)
	})
}

// registerLevelRoutes exposes the read once per level. The level is part of the path rather than a
// query value, so that a caller holding org-level permission cannot reach the tenant level by
// sending a different one.
//
// There is no write route here. A module writes its own settings at
// `PATCH /v1/{module_key}/settings/:level`, where it can validate against its own data first, and
// persists them through core's settings port; a generic write here would let a caller skip that.
func registerLevelRoutes(routeV1 *echo.Group, settingsRest *v1.SettingsRest) {
	routeV1.GET("/tenant/:module_key", settingsRest.GetTenantSettings, m.UserSmokeAuthz())

	routeV1.GET("/org/:module_key", settingsRest.GetOrgSettings, m.UserSmokeAuthz())

	routeV1.GET("/user/:module_key", settingsRest.GetUserPreferences, m.UserSmokeAuthz())
}

// registerReadOnly serves a settings resource's reads only: Settings never takes a generic write
// (see registerLevelRoutes).
func registerReadOnly(route *echo.Group, schemaName string, handlers composable.CrudRestHandlers) error {
	return composable.NewRestEngine(schemaName, handlers).
		AddCrudRoutes(dynamicengines.ReadOnlyCrudActions()...).
		RegisterRoutes(route)
}
