package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

// initMeV1 serves the caller's own context. Authenticated-only: it answers only about the caller's
// own access, and gating it behind an entitlement would keep it from the plain users who need it
// to understand a refusal.
func initMeV1() error {
	return deps.Invoke(func(route *echo.Group, userRest *v1.UserRest, permissionRest *v1.PermissionRest) {
		routeV1 := route.Group("/v1/iam")
		routeV1.GET("/me/context", userRest.GetUserContext, m.UserSmokeAuthz())
		routeV1.POST("/me/test-permissions", permissionRest.TestMyPermissions, m.UserSmokeAuthz())
	})
}

func initUserCrudV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.UserCrudRest) error {
		return userRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func userRestEngine(rest *v1.UserCrudRest) dynamicresource.RestEngine {
	return dynamicresource.NewRestEngine(models.UserSchemaName, rest).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/send_invitation", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.SendInvitation,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_roles", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageRoles,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/roles", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.Roles,
		})
}
