package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
)

func NewAttributeHandler(attributeSvc itAttribute.AttributeService) *AttributeHandler {
	return &AttributeHandler{
		AttributeSvc: attributeSvc,
	}
}

type AttributeHandler struct {
	AttributeSvc itAttribute.AttributeService
}

func (this *AttributeHandler) CreateAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.CreateAttributeCommand]) (*cqrs.Reply[itAttribute.CreateAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.CreateAttribute)
}

func (this *AttributeHandler) UpdateAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.UpdateAttributeCommand]) (*cqrs.Reply[itAttribute.UpdateAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.UpdateAttribute)
}

func (this *AttributeHandler) DeleteAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.DeleteAttributeCommand]) (*cqrs.Reply[itAttribute.DeleteAttributeResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.DeleteAttribute)
}

func (this *AttributeHandler) GetAttributeById(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.GetAttributeByIdQuery]) (*cqrs.Reply[itAttribute.GetAttributeByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.GetAttributeById)
}

func (this *AttributeHandler) SearchAttributes(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.SearchAttributesQuery]) (*cqrs.Reply[itAttribute.SearchAttributesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeSvc.SearchAttributes)
}
