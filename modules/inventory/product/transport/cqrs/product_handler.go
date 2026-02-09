package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewProductHandler(productSvc itProduct.ProductService) *ProductHandler {
	return &ProductHandler{
		ProductSvc: productSvc,
	}
}

type ProductHandler struct {
	ProductSvc itProduct.ProductService
}

func (this *ProductHandler) CreateProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.CreateProductCommand]) (*cqrs.Reply[itProduct.CreateProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.CreateProduct)
}

func (this *ProductHandler) UpdateProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.UpdateProductCommand]) (*cqrs.Reply[itProduct.UpdateProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.UpdateProduct)
}

func (this *ProductHandler) DeleteProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.DeleteProductCommand]) (*cqrs.Reply[itProduct.DeleteProductResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.DeleteProduct)
}

func (this *ProductHandler) GetProductById(ctx context.Context, packet *cqrs.RequestPacket[itProduct.GetProductByIdQuery]) (*cqrs.Reply[itProduct.GetProductByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.GetProductById)
}

func (this *ProductHandler) SearchProducts(ctx context.Context, packet *cqrs.RequestPacket[itProduct.SearchProductsQuery]) (*cqrs.Reply[itProduct.SearchProductsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ProductSvc.SearchProducts)
}
