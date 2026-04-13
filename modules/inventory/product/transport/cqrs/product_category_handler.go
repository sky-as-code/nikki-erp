package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

func NewProductCategoryHandler(productCategorySvc itProductCategory.ProductCategoryService, logger logging.LoggerService) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		Logger:             logger,
		ProductCategorySvc: productCategorySvc,
	}
}

type ProductCategoryHandler struct {
	Logger             logging.LoggerService
	ProductCategorySvc itProductCategory.ProductCategoryService
}

func (this *ProductCategoryHandler) CreateProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.CreateProductCategoryCommand]) (
	*cqrs.Reply[itProductCategory.CreateProductCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductCategorySvc.CreateProductCategory)
}

func (this *ProductCategoryHandler) UpdateProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.UpdateProductCategoryCommand]) (
	*cqrs.Reply[itProductCategory.UpdateProductCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductCategorySvc.UpdateProductCategory)
}

func (this *ProductCategoryHandler) DeleteProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.DeleteProductCategoryCommand]) (
	*cqrs.Reply[itProductCategory.DeleteProductCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductCategorySvc.DeleteProductCategory)
}

func (this *ProductCategoryHandler) GetProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.GetProductCategoryQuery]) (
	*cqrs.Reply[itProductCategory.GetProductCategoryResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductCategorySvc.GetProductCategory)
}

func (this *ProductCategoryHandler) SearchProductCategories(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.SearchProductCategoriesQuery]) (
	*cqrs.Reply[itProductCategory.SearchProductCategoriesResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductCategorySvc.SearchProductCategories)
}
