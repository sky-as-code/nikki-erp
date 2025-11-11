package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewUnitCategoryRest,
	)
	return deps.Invoke(func(route *echo.Group, unitCategoryRest *v1.UnitCategoryRest) {
		v1 := route.Group("/v1/inventory")
		initV1(v1, unitCategoryRest)
	})
}

func initV1(route *echo.Group, unitCategoryRest *v1.UnitCategoryRest) {
	route.POST("/unit-categories", unitCategoryRest.CreateUnitCategory)
	route.DELETE("/unit-categories/:id", unitCategoryRest.DeleteUnitCategory)
	route.GET("/unit-categories/:id", unitCategoryRest.GetUnitCategoryById)
	route.GET("/unit-categories", unitCategoryRest.SearchUnitCategories)
	route.PUT("/unit-categories/:id", unitCategoryRest.UpdateUnitCategory)
}
