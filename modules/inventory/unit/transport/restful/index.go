package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewUnitRest,
		v1.NewUnitCategoryRest,
	)
	return stdErr.Join(err, initUnitV1())
}

func initUnitV1() error {
	return deps.Invoke(func(
		route *echo.Group,
	) error {
		routeV1 := route.Group("/v1/:org_id/inventory")
		routeV1.Use(middlewares.RequestContextMiddleware2(constants.InventoryModuleName))

		err := stdErr.Join(
			httpserver.RegisterBasicCrudRest[*v1.UnitRest]("/units", routeV1),
			httpserver.RegisterBasicCrudRest[*v1.UnitCategoryRest]("/unit-categories", routeV1),
		)
		return err
	})
}
