package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initGroupV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.GroupCrudRest) error {
		return groupRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func groupRestEngine(rest *v1.GroupCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.GroupSchemaName, rest).
		AddCrudRoutes().
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_users", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageUsers,
		}).
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_roles", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageRoles,
		}).
		AddRoute(composable.RouteDefinition{
			Path: ":id/roles", ActionType: composable.ActionTypeRead, HandlerFn: rest.Roles,
		})
}
