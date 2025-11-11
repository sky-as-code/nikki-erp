package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/variant/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewVariantRest,
	)
	return deps.Invoke(func(route *echo.Group, variantRest *v1.VariantRest) {
		v1 := route.Group("/v1/inventory")
		initV1(v1, variantRest)
	})
}

func initV1(route *echo.Group, variantRest *v1.VariantRest) {
	route.POST("/variants", variantRest.CreateVariant)
	route.DELETE("/variants/:id", variantRest.DeleteVariant)
	route.GET("/variants/:id", variantRest.GetVariantById)
	route.GET("/variants", variantRest.SearchVariants)
	route.PUT("/variants/:id", variantRest.UpdateVariant)
}
