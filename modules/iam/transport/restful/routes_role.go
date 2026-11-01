package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initRoleV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.RoleCrudRest) error {
		return roleRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func roleRestEngine(rest *v1.RoleCrudRest) dynamicresource.RestEngine {
	return dynamicresource.NewRestEngine(models.RoleSchemaName, rest).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_users", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageUsers,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_services", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageServices,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_groups", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageGroups,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: "describe", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.Describe,
		})
}
