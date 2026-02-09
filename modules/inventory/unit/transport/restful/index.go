package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewUnitRest,
		v1.NewUnitCategoryRest,
	)
	return deps.Invoke(func(
		route *echo.Group,
		unitRest *v1.UnitRest,
		unitCategoryRest *v1.UnitCategoryRest,
	) {
		unitV1 := route.Group("/v1/:orgId/inventory")
		initV1(unitV1, unitRest, unitCategoryRest)
	})
}

func initV1(route *echo.Group, unitRest *v1.UnitRest, unitCategoryRest *v1.UnitCategoryRest) {
	// protected := route.Group("", commonMiddleware.RequireAuthMiddleware())

	// Unit routes
	route.POST("/units", unitRest.CreateUnit)
	route.PUT("/units/:id", unitRest.UpdateUnit)
	route.DELETE("/units/:id", unitRest.DeleteUnit)
	route.GET("/units/:id", unitRest.GetUnitById)
	route.GET("/units", unitRest.SearchUnits)

	// UnitCategory routes
	route.POST("/unit-categories", unitCategoryRest.CreateUnitCategory)
	route.PUT("/unit-categories/:id", unitCategoryRest.UpdateUnitCategory)
	route.DELETE("/unit-categories/:id", unitCategoryRest.DeleteUnitCategory)
	route.GET("/unit-categories/:id", unitCategoryRest.GetUnitCategoryById)
	route.GET("/unit-categories", unitCategoryRest.SearchUnitCategories)
}
