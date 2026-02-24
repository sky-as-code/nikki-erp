package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

func NewProductCategoryHandler(productCategorySvc itProductCategory.ProductCategoryService) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		ProductCategorySvc: productCategorySvc,
	}
}

type ProductCategoryHandler struct {
	ProductCategorySvc itProductCategory.ProductCategoryService
}

func (this *ProductCategoryHandler) CreateProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.CreateProductCategoryCommand]) (*cqrs.Reply[itProductCategory.CreateProductCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductCategorySvc.CreateProductCategory)
}

func (this *ProductCategoryHandler) UpdateProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.UpdateProductCategoryCommand]) (*cqrs.Reply[itProductCategory.UpdateProductCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductCategorySvc.UpdateProductCategory)
}

func (this *ProductCategoryHandler) DeleteProductCategory(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.DeleteProductCategoryCommand]) (*cqrs.Reply[itProductCategory.DeleteProductCategoryResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductCategorySvc.DeleteProductCategory)
}

func (this *ProductCategoryHandler) GetProductCategoryById(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.GetProductCategoryByIdQuery]) (*cqrs.Reply[itProductCategory.GetProductCategoryByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductCategorySvc.GetProductCategoryById)
}

func (this *ProductCategoryHandler) SearchProductCategories(ctx context.Context, packet *cqrs.RequestPacket[itProductCategory.SearchProductCategoriesQuery]) (*cqrs.Reply[itProductCategory.SearchProductCategoriesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductCategorySvc.SearchProductCategories)
}
