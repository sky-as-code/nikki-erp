package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initServiceCrudV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.ServiceCrudRest) error {
		return serviceRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func serviceRestEngine(rest *v1.ServiceCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.ServiceSchemaName, rest).
		AddCrudRoutes().
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_roles", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageRoles,
		}).
		AddRoute(composable.RouteDefinition{
			Path: ":id/roles", ActionType: composable.ActionTypeRead, HandlerFn: rest.Roles,
		})
}
