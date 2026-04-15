package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewModuleRest,
		v1.NewUnitRest,
		v1.NewUnitCategoryRest,
	)
	return stdErr.Join(
		err,
		initEssentialV1(),
		initUnitV1(),
	)
}

func initEssentialV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		moduleRest *v1.ModuleRest,
	) {
		routeV1 := route.Group("/v1/essential")

		routeV1.DELETE("/modules/:id", moduleRest.DeleteModule)
		routeV1.GET("/modules/:id", moduleRest.GetModule)
		routeV1.GET("/modules", moduleRest.SearchModules)
		routeV1.POST("/modules/exists", moduleRest.ModuleExists)
		routeV1.POST("/modules", moduleRest.CreateModule)
		routeV1.PUT("/modules/:id", moduleRest.UpdateModule)
	})
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
