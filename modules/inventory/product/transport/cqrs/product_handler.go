package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces"
)

func NewProductHandler(productSvc it.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductSvc: productSvc,
	}
}

type ProductHandler struct {
	ProductSvc it.ProductService
}

func (this *ProductHandler) CreateProduct(ctx context.Context, packet *cqrs.RequestPacket[it.CreateProductCommand]) (*cqrs.Reply[it.CreateProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.CreateProduct)
}

func (this *ProductHandler) UpdateProduct(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateProductCommand]) (*cqrs.Reply[it.UpdateProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.UpdateProduct)
}

func (this *ProductHandler) DeleteProduct(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteProductCommand]) (*cqrs.Reply[it.DeleteProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.DeleteProduct)
}

func (this *ProductHandler) GetProductById(ctx context.Context, packet *cqrs.RequestPacket[it.GetProductByIdQuery]) (*cqrs.Reply[it.GetProductByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.GetProductById)
}

func (this *ProductHandler) SearchProducts(ctx context.Context, packet *cqrs.RequestPacket[it.SearchProductsQuery]) (*cqrs.Reply[it.SearchProductsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.SearchProducts)
}
