package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initGrantRequestV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.GrantRequestCrudRest) error {
		return grantRequestRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func grantRequestRestEngine(rest *v1.GrantRequestCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.RoleRequestSchemaName, rest).AddCrudRoutes()
}
