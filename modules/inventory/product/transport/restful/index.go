package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/product/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewAttributeRest,
		v1.NewAttributeGroupRest,
		v1.NewAttributeValueRest,
		v1.NewProductRest,
		v1.NewProductCategoryRest,
		v1.NewVariantRest,
	)
	return stdErr.Join(err, initProductV1())
}

func initProductV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		attributeRest *v1.AttributeRest,
		attributeGroupRest *v1.AttributeGroupRest,
		attributeValueRest *v1.AttributeValueRest,
		productRest *v1.ProductRest,
		productCategoryRest *v1.ProductCategoryRest,
		variantRest *v1.VariantRest,
	) error {
		routeV1 := route.Group("/v1/:org_id/inventory")

		routeV1.DELETE("/products-categories/:id", productCategoryRest.Delete)
		routeV1.GET("/products-categories/:id", productCategoryRest.GetOne)
		routeV1.POST("/products-categories/:id/exists", productCategoryRest.Exists)
		routeV1.POST("/products-categories/:id", productCategoryRest.Create)
		routeV1.PUT("/products-categories/:id", productCategoryRest.Update)

		routeV1.DELETE("/products/:id", productRest.Delete)
		routeV1.GET("/products/:id", productRest.GetOne)
		routeV1.POST("/products/:id/exists", productRest.Exists)
		routeV1.POST("/products/:id/archived", productRest.SetIsArchived)
		routeV1.POST("/products/:id", productRest.Create)
		routeV1.PUT("/products/:id", productRest.Update)

		// Register nested resource routes manually
		// Attributes (nested under products)
		routeV1.DELETE("/products/:product_id/attributes/:id", attributeRest.Delete)
		routeV1.GET("/products/:product_id/attributes", attributeRest.Search)
		routeV1.GET("/products/:product_id/attributes/:id", attributeRest.GetOne)
		routeV1.POST("/products/:product_id/attributes/exists", attributeRest.Exists)
		routeV1.POST("/products/:product_id/attributes", attributeRest.Create)
		routeV1.PUT("/products/:product_id/attributes/:id", attributeRest.Update)

		// Attribute Groups (nested under products)
		routeV1.DELETE("/products/:product_id/attribute-groups/:id", attributeGroupRest.Delete)
		routeV1.GET("/products/:product_id/attribute-groups/:id", attributeGroupRest.GetOne)
		routeV1.GET("/products/:product_id/attribute-groups", attributeGroupRest.Search)
		routeV1.POST("/products/:product_id/attribute-groups/exists", attributeGroupRest.Exists)
		routeV1.POST("/products/:product_id/attribute-groups", attributeGroupRest.Create)
		routeV1.PUT("/products/:product_id/attribute-groups/:id", attributeGroupRest.Update)

		// Attribute Values (nested under products and attributes)
		routeV1.DELETE("/products/:product_id/attributes/:attribute_id/values/:id", attributeValueRest.Delete)
		routeV1.GET("/products/:product_id/attributes/:attribute_id/values", attributeValueRest.Search)

		// Variants (nested under products, plus standalone search)
		routeV1.DELETE("/products/:product_id/variants/:id", variantRest.Delete)
		routeV1.GET("/products/:product_id/variants", variantRest.Search)
		routeV1.GET("/products/:product_id/variants/:id", variantRest.GetOne)
		routeV1.POST("/products/:product_id/variants/exists", variantRest.Exists)
		routeV1.POST("/products/:product_id/variants", variantRest.Create)
		routeV1.PUT("/products/:product_id/variants/:id", variantRest.Update)

		// Standalone variant search
		routeV1.GET("/variants", variantRest.Search)

		return nil
	})
}
