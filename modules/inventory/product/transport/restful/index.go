package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/product/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewAttributeRest,
		v1.NewAttributeGroupRest,
		v1.NewAttributeValueRest,
		v1.NewProductRest,
		v1.NewVariantRest,
		v1.NewProductCategoryRest,
	)
	return deps.Invoke(func(
		route *echo.Group,
		attributeRest *v1.AttributeRest,
		attributeGroupRest *v1.AttributeGroupRest,
		attributeValueRest *v1.AttributeValueRest,
		productRest *v1.ProductRest,
		variantRest *v1.VariantRest,
		productCategoryRest *v1.ProductCategoryRest,
	) {
		v1 := route.Group("/v1/:orgId/inventory")
		initV1(v1, attributeRest, attributeGroupRest, attributeValueRest, productRest, variantRest, productCategoryRest)
	})
}

func initV1(
	route *echo.Group,
	attributeRest *v1.AttributeRest,
	attributeGroupRest *v1.AttributeGroupRest,
	attributeValueRest *v1.AttributeValueRest,
	productRest *v1.ProductRest,
	variantRest *v1.VariantRest,
	productCategoryRest *v1.ProductCategoryRest,
) {
	route.POST("/products/:productId/attributes", attributeRest.CreateAttribute)
	route.PUT("/products/:productId/attributes/:id", attributeRest.UpdateAttribute)
	route.DELETE("/products/:productId/attributes/:id", attributeRest.DeleteAttribute)
	route.GET("/products/:productId/attributes/:id", attributeRest.GetAttributeById)
	route.GET("/products/:productId/attributes", attributeRest.SearchAttributes)

	route.POST("/products/:productId/attribute-group", attributeGroupRest.CreateAttributeGroup)
	route.PUT("/products/:productId/attribute-group/:id", attributeGroupRest.UpdateAttributeGroup)
	route.DELETE("/products/:productId/attribute-group/:id", attributeGroupRest.DeleteAttributeGroup)
	route.GET("/products/:productId/attribute-group/:id", attributeGroupRest.GetAttributeGroupById)
	route.GET("/products/:productId/attribute-group", attributeGroupRest.SearchAttributeGroups)

	// AttributeValue routes
	// route.POST("/attribute-values", attributeValueRest.CreateAttributeValue)
	// route.PUT("/attribute-values/:id", attributeValueRest.UpdateAttributeValue)
	route.DELETE("/products/:productId/attributes/:attributeId/values/:id", attributeValueRest.DeleteAttributeValue)
	// route.GET("/attribute-values/:id", attributeValueRest.GetAttributeValueById)
	route.GET("/products/:productId/attributes/:attributeId/values", attributeValueRest.SearchAttributeValues)

	// Product routes
	route.POST("/products", productRest.CreateProduct)
	route.PUT("/products/:id", productRest.UpdateProduct)
	route.DELETE("/products/:id", productRest.DeleteProduct)
	route.GET("/products/:id", productRest.GetProductById)
	route.GET("/products", productRest.SearchProducts)

	route.POST("/products-category", productCategoryRest.CreateProductCategory)
	route.PUT("/products-category/:id", productCategoryRest.UpdateProductCategory)
	route.DELETE("/products-category/:id", productCategoryRest.DeleteProductCategory)
	route.GET("/products-category/:id", productCategoryRest.GetProductCategoryById)
	route.GET("/products-category", productCategoryRest.SearchProductCategories)

	// Variant routes
	route.POST("/products/:productId/variants", variantRest.CreateVariant)
	route.PUT("/products/:productId/variants/:id", variantRest.UpdateVariant)
	route.DELETE("/products/:productId/variants/:id", variantRest.DeleteVariant)
	route.GET("/products/:productId/variants/:id", variantRest.GetVariantById)
	route.GET("/products/:productId/variants", variantRest.SearchVariants)
}
