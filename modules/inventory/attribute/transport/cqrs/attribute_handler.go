package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/interfaces"
)

func NewAttributeHandler(attributeSvc it.AttributeService) *AttributeHandler {
	return &AttributeHandler{
		AttributeSvc: attributeSvc,
	}
}

type AttributeHandler struct {
	AttributeSvc it.AttributeService
}

func (this *AttributeHandler) CreateAttribute(ctx context.Context, packet *cqrs.RequestPacket[it.CreateAttributeCommand]) (*cqrs.Reply[it.CreateAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.CreateAttribute)
}

func (this *AttributeHandler) UpdateAttribute(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateAttributeCommand]) (*cqrs.Reply[it.UpdateAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.UpdateAttribute)
}

func (this *AttributeHandler) DeleteAttribute(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteAttributeCommand]) (*cqrs.Reply[it.DeleteAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.DeleteAttribute)
}

func (this *AttributeHandler) GetAttributeById(ctx context.Context, packet *cqrs.RequestPacket[it.GetAttributeByIdQuery]) (*cqrs.Reply[it.GetAttributeByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.GetAttributeById)
}

func (this *AttributeHandler) SearchAttributes(ctx context.Context, packet *cqrs.RequestPacket[it.SearchAttributesQuery]) (*cqrs.Reply[it.SearchAttributesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.SearchAttributes)
}
