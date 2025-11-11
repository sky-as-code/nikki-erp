package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewAttributeRest,
	)
	return deps.Invoke(func(route *echo.Group, attributeRest *v1.AttributeRest) {
		v1 := route.Group("/v1/:orgId/inventory/products/:productId")
		initV1(v1, attributeRest)
	})
}

func initV1(route *echo.Group, attributeRest *v1.AttributeRest) {
	route.POST("/attributes", attributeRest.CreateAttribute)
	route.DELETE("/attributes/:id", attributeRest.DeleteAttribute)
	route.GET("/attributes/:id", attributeRest.GetAttributeById)
	route.GET("/attributes", attributeRest.SearchAttributes)
	route.PUT("/attributes/:id", attributeRest.UpdateAttribute)
}
