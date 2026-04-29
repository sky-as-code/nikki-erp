package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/settings/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewUserPreferenceRest,
	)
	return stdErr.Join(err, initSettingsV1())
}

func initSettingsV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		userPreferenceRest *v1.UserPreferenceRest,
	) error {
		routeV1 := route.Group("/v1/settings")

		routeV1.DELETE("/user-preferences/:id", userPreferenceRest.DeleteUserPreference)
		routeV1.GET("/user-preferences/meta/schema", userPreferenceRest.GetModelSchema)
		routeV1.GET("/user-preferences/:id", userPreferenceRest.GetUserPreference)
		routeV1.GET("/user-preferences", userPreferenceRest.SearchUserPreferences)
		routeV1.POST("/user-preferences/exists", userPreferenceRest.UserPreferenceExists)
		routeV1.POST("/user-preferences", userPreferenceRest.CreateUserPreference)
		routeV1.PATCH("/user-preferences/:id", userPreferenceRest.UpdateUserPreference)

		return nil
	})
}
