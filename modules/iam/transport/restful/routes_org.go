package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initOrgV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.OrgCrudRest) error {
		return orgRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func orgRestEngine(rest *v1.OrgCrudRest) dynamicresource.RestEngine {
	return dynamicresource.NewRestEngine(models.OrganizationSchemaName, rest).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_users", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageUsers,
		})
}
