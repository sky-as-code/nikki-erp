package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func initResourceV1() error {
	return deps.Invoke(func(route *echo.Group, resourceRest *v1.ResourceCrudRest, actionRest *v1.ActionCrudRest) error {
		routeV1 := route.Group("/v1/iam")
		return stdErr.Join(
			resourceRestEngine(resourceRest).RegisterRoutes(routeV1),
			actionRestEngine(actionRest).RegisterRoutes(routeV1),
		)
	})
}

func resourceRestEngine(rest *v1.ResourceCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.ResourceSchemaName, rest).AddCrudRoutes()
}

func actionRestEngine(rest *v1.ActionCrudRest) composable.RestEngine {
	return composable.NewRestEngine(models.ActionSchemaName, rest).AddCrudRoutes()
}
