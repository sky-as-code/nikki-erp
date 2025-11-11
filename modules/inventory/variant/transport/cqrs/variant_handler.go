package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

func NewVariantHandler(variantSvc it.VariantService) *VariantHandler {
	return &VariantHandler{
		VariantSvc: variantSvc,
	}
}

type VariantHandler struct {
	VariantSvc it.VariantService
}

func (this *VariantHandler) CreateVariant(ctx context.Context, packet *cqrs.RequestPacket[it.CreateVariantCommand]) (*cqrs.Reply[it.CreateVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.CreateVariant)
}

func (this *VariantHandler) UpdateVariant(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateVariantCommand]) (*cqrs.Reply[it.UpdateVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.UpdateVariant)
}

func (this *VariantHandler) DeleteVariant(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteVariantCommand]) (*cqrs.Reply[it.DeleteVariantResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.DeleteVariant)
}

func (this *VariantHandler) GetVariantById(ctx context.Context, packet *cqrs.RequestPacket[it.GetVariantByIdQuery]) (*cqrs.Reply[it.GetVariantByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.GetVariantById)
}

func (this *VariantHandler) SearchVariants(ctx context.Context, packet *cqrs.RequestPacket[it.SearchVariantsQuery]) (*cqrs.Reply[it.SearchVariantsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.VariantSvc.SearchVariants)
}
