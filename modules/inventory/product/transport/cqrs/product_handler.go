package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewProductHandler(productSvc itProduct.ProductService, logger logging.LoggerService) *ProductHandler {
	return &ProductHandler{
		Logger:     logger,
		ProductSvc: productSvc,
	}
}

type ProductHandler struct {
	Logger     logging.LoggerService
	ProductSvc itProduct.ProductService
}

func (this *ProductHandler) CreateProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.CreateProductCommand]) (
	*cqrs.Reply[itProduct.CreateProductResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.CreateProduct)
}

func (this *ProductHandler) UpdateProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.UpdateProductCommand]) (
	*cqrs.Reply[itProduct.UpdateProductResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.UpdateProduct)
}

func (this *ProductHandler) DeleteProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.DeleteProductCommand]) (
	*cqrs.Reply[itProduct.DeleteProductResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.DeleteProduct)
}

func (this *ProductHandler) GetProduct(ctx context.Context, packet *cqrs.RequestPacket[itProduct.GetProductQuery]) (
	*cqrs.Reply[itProduct.GetProductResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.GetProduct)
}

func (this *ProductHandler) ProductExists(ctx context.Context, packet *cqrs.RequestPacket[itProduct.ProductExistsQuery]) (
	*cqrs.Reply[itProduct.ProductExistsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.ProductExists)
}

func (this *ProductHandler) SetProductIsArchived(ctx context.Context, packet *cqrs.RequestPacket[itProduct.SetProductIsArchivedCommand]) (
	*cqrs.Reply[itProduct.SetProductIsArchivedResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.SetProductIsArchived)
}

func (this *ProductHandler) SearchProducts(ctx context.Context, packet *cqrs.RequestPacket[itProduct.SearchProductsQuery]) (
	*cqrs.Reply[itProduct.SearchProductsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.ProductSvc.SearchProducts)
}
