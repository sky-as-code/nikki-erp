package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantHandler(variantSvc itVariant.VariantService, logger logging.LoggerService) *VariantHandler {
	return &VariantHandler{
		Logger:     logger,
		VariantSvc: variantSvc,
	}
}

type VariantHandler struct {
	Logger     logging.LoggerService
	VariantSvc itVariant.VariantService
}

func (this *VariantHandler) CreateVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.CreateVariantCommand]) (
	*cqrs.Reply[itVariant.CreateVariantResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.VariantSvc.CreateVariant)
}

func (this *VariantHandler) UpdateVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.UpdateVariantCommand]) (
	*cqrs.Reply[itVariant.UpdateVariantResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.VariantSvc.UpdateVariant)
}

func (this *VariantHandler) DeleteVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.DeleteVariantCommand]) (
	*cqrs.Reply[itVariant.DeleteVariantResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.VariantSvc.DeleteVariant)
}

func (this *VariantHandler) GetVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.GetVariantQuery]) (
	*cqrs.Reply[itVariant.GetVariantResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.VariantSvc.GetVariant)
}

func (this *VariantHandler) SearchVariants(ctx context.Context, packet *cqrs.RequestPacket[itVariant.SearchVariantsQuery]) (
	*cqrs.Reply[itVariant.SearchVariantsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.VariantSvc.SearchVariants)
}
