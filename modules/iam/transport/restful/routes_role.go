package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initRoleV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.RoleCrudRest) error {
		return roleRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func roleRestEngine(rest *v1.RoleCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.RoleSchemaName, rest).
		AddCrudRoutes().
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_users", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageUsers,
		}).
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_groups", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageGroups,
		}).
		AddRoute(composable.RouteDefinition{
			Path: "describe", ActionType: composable.ActionTypeRead, HandlerFn: rest.Describe,
		})
}
