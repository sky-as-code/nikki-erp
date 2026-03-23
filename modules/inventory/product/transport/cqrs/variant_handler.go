package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantHandler(variantSvc itVariant.VariantService) *VariantHandler {
	return &VariantHandler{
		VariantSvc: variantSvc,
	}
}

type VariantHandler struct {
	VariantSvc itVariant.VariantService
}

func (this *VariantHandler) CreateVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.CreateVariantCommand]) (*cqrs.Reply[itVariant.CreateVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.CreateVariant)
}

func (this *VariantHandler) UpdateVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.UpdateVariantCommand]) (*cqrs.Reply[itVariant.UpdateVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.UpdateVariant)
}

func (this *VariantHandler) DeleteVariant(ctx context.Context, packet *cqrs.RequestPacket[itVariant.DeleteVariantCommand]) (*cqrs.Reply[itVariant.DeleteVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.DeleteVariant)
}

func (this *VariantHandler) GetVariantById(ctx context.Context, packet *cqrs.RequestPacket[itVariant.GetVariantByIdQuery]) (*cqrs.Reply[itVariant.GetVariantByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.GetVariantById)
}

func (this *VariantHandler) SearchVariants(ctx context.Context, packet *cqrs.RequestPacket[itVariant.SearchVariantsQuery]) (*cqrs.Reply[itVariant.SearchVariantsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.SearchVariants)
}
