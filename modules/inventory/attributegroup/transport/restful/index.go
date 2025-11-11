package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewAttributeGroupRest,
	)
	return deps.Invoke(func(route *echo.Group, attributeGroupRest *v1.AttributeGroupRest) {
		v1 := route.Group("/v1/:orgId/inventory")
		initV1(v1, attributeGroupRest)
	})
}

func initV1(route *echo.Group, attributeGroupRest *v1.AttributeGroupRest) {
	route.POST("/attribute-groups", attributeGroupRest.CreateAttributeGroup)
	route.PUT("/attribute-groups/:id", attributeGroupRest.UpdateAttributeGroup)
	route.DELETE("/attribute-groups/:id", attributeGroupRest.DeleteAttributeGroup)
	route.GET("/attribute-groups/:id", attributeGroupRest.GetAttributeGroupById)
	route.GET("/attribute-groups", attributeGroupRest.SearchAttributeGroups)
}
