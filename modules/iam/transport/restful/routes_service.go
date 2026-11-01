package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initServiceCrudV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.ServiceCrudRest) error {
		return serviceRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func serviceRestEngine(rest *v1.ServiceCrudRest) dynamicresource.RestEngine {
	return dynamicresource.NewRestEngine(models.ServiceSchemaName, rest).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_roles", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageRoles,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/roles", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.Roles,
		})
}
