package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewUnitRest,
	)
	return deps.Invoke(func(route *echo.Group, unitRest *v1.UnitRest) {
		v1 := route.Group("/v1/:orgId/inventory")
		initV1(v1, unitRest)
	})
}

func initV1(route *echo.Group, unitRest *v1.UnitRest) {
	route.POST("/units", unitRest.CreateUnit)
	route.DELETE("/units/:id", unitRest.DeleteUnit)
	route.GET("/units/:id", unitRest.GetUnitById)
	route.GET("/units", unitRest.SearchUnits)
	route.PUT("/units/:id", unitRest.UpdateUnit)
}
