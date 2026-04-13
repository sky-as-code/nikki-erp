package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/product/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewAttributeRest,
		v1.NewAttributeGroupRest,
		v1.NewAttributeValueRest,
		v1.NewProductRest,
		v1.NewVariantRest,
		v1.NewProductCategoryRest,
	)
	return stdErr.Join(err, initProductV1())
}

func initProductV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		attributeRest *v1.AttributeRest,
		attributeGroupRest *v1.AttributeGroupRest,
		attributeValueRest *v1.AttributeValueRest,
		variantRest *v1.VariantRest,
	) error {
		routeV1 := route.Group("/v1/:org_id/inventory")
		routeV1.Use(middlewares.RequestContextMiddleware2(constants.InventoryModuleName))

		// Register standard CRUD resources using helpers
		err := stdErr.Join(
			httpserver.RegisterArchivableCrudRest[*v1.ProductRest]("/products", routeV1),
			httpserver.RegisterBasicCrudRest[*v1.ProductCategoryRest]("/product-categories", routeV1),
		)
		if err != nil {
			return err
		}

		// Register nested resource routes manually
		// Attributes (nested under products)
		routeV1.POST("/products/:product_id/attributes", attributeRest.Create)
		routeV1.PUT("/products/:product_id/attributes/:id", attributeRest.Update)
		routeV1.DELETE("/products/:product_id/attributes/:id", attributeRest.Delete)
		routeV1.GET("/products/:product_id/attributes/:id", attributeRest.GetOne)
		routeV1.GET("/products/:product_id/attributes", attributeRest.Search)
		routeV1.POST("/products/:product_id/attributes/exists", attributeRest.Exists)

		// Attribute Groups (nested under products)
		routeV1.POST("/products/:product_id/attribute-groups", attributeGroupRest.Create)
		routeV1.PUT("/products/:product_id/attribute-groups/:id", attributeGroupRest.Update)
		routeV1.DELETE("/products/:product_id/attribute-groups/:id", attributeGroupRest.Delete)
		routeV1.GET("/products/:product_id/attribute-groups/:id", attributeGroupRest.GetOne)
		routeV1.GET("/products/:product_id/attribute-groups", attributeGroupRest.Search)
		routeV1.POST("/products/:product_id/attribute-groups/exists", attributeGroupRest.Exists)

		// Attribute Values (nested under products and attributes)
		routeV1.DELETE("/products/:product_id/attributes/:attribute_id/values/:id", attributeValueRest.Delete)
		routeV1.GET("/products/:product_id/attributes/:attribute_id/values", attributeValueRest.Search)

		// Variants (nested under products, plus standalone search)
		routeV1.POST("/products/:product_id/variants", variantRest.Create)
		routeV1.PUT("/products/:product_id/variants/:id", variantRest.Update)
		routeV1.DELETE("/products/:product_id/variants/:id", variantRest.Delete)
		routeV1.GET("/products/:product_id/variants/:id", variantRest.GetOne)
		routeV1.GET("/products/:product_id/variants", variantRest.Search)
		routeV1.POST("/products/:product_id/variants/exists", variantRest.Exists)

		// Standalone variant search
		routeV1.GET("/variants", variantRest.Search)

		return nil
	})
}
