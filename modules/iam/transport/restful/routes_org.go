package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initOrgV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.OrgCrudRest) error {
		return orgRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func orgRestEngine(rest *v1.OrgCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.OrganizationSchemaName, rest).
		AddCrudRoutes().
		AddRoute(composable.RouteDefinition{
			Path: ":id/manage_users", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ManageUsers,
		})
}
