package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewAttributeValueRest,
	)
	return deps.Invoke(func(route *echo.Group, attributeValueRest *v1.AttributeValueRest) {
		v1 := route.Group("/v1/inventory")
		initV1(v1, attributeValueRest)
	})
}

func initV1(route *echo.Group, attributeValueRest *v1.AttributeValueRest) {
	route.POST("/attribute-values", attributeValueRest.CreateAttributeValue)
	route.DELETE("/attribute-values/:id", attributeValueRest.DeleteAttributeValue)
	route.GET("/attribute-values/:id", attributeValueRest.GetAttributeValueById)
	route.GET("/attribute-values", attributeValueRest.SearchAttributeValues)
	route.PUT("/attribute-values/:id", attributeValueRest.UpdateAttributeValue)
}
