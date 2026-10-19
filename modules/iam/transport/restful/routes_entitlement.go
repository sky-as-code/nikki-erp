package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initEntitlementV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.EntitlementCrudRest) error {
		return entitlementRestEngine(rest).RegisterRoutes(route.Group("/v1/iam"))
	})
}

func entitlementRestEngine(rest *v1.EntitlementCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.EntitlementSchemaName, rest).AddCrudRoutes()
}
