package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/product/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewProductRest,
	)
	return deps.Invoke(func(route *echo.Group, productRest *v1.ProductRest) {
		v1 := route.Group("/v1/:orgId/inventory")
		initV1(v1, productRest)
	})
}

func initV1(route *echo.Group, productRest *v1.ProductRest) {
	route.POST("/products", productRest.CreateProduct)
	route.DELETE("/products/:id", productRest.DeleteProduct)
	route.GET("/products/:id", productRest.GetProductById)
	route.GET("/products", productRest.SearchProducts)
	route.PUT("/products/:id", productRest.UpdateProduct)
}
