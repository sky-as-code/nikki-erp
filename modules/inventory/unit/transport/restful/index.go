package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
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
		unitRest *v1.UnitRest,
		unitCategoryRest *v1.UnitCategoryRest,
	) {
		routeV1 := route.Group("/v1/:org_id/inventory")

		routeV1.DELETE("/units/:id", unitRest.Delete)
		routeV1.GET("/units/:id", unitRest.GetOne)
		routeV1.POST("/units/:id/exists", unitRest.Exists)
		routeV1.POST("/units/:id", unitRest.Create)
		routeV1.PUT("/units/:id", unitRest.Update)

		routeV1.DELETE("/units-categories/:id", unitCategoryRest.Delete)
		routeV1.GET("/units-categories/:id", unitCategoryRest.GetOne)
		routeV1.POST("/units-categories/:id/exists", unitCategoryRest.Exists)
		routeV1.POST("/units-categories/:id", unitCategoryRest.Create)
		routeV1.PUT("/units-categories/:id", unitCategoryRest.Update)
	})
}
