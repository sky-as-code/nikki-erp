package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
)

func NewAttributeHandler(attributeSvc itAttribute.AttributeService, logger logging.LoggerService) *AttributeHandler {
	return &AttributeHandler{
		Logger:       logger,
		AttributeSvc: attributeSvc,
	}
}

type AttributeHandler struct {
	Logger       logging.LoggerService
	AttributeSvc itAttribute.AttributeService
}

func (this *AttributeHandler) CreateAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.CreateAttributeCommand]) (
	*cqrs.Reply[itAttribute.CreateAttributeResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeSvc.CreateAttribute)
}

func (this *AttributeHandler) UpdateAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.UpdateAttributeCommand]) (
	*cqrs.Reply[itAttribute.UpdateAttributeResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeSvc.UpdateAttribute)
}

func (this *AttributeHandler) DeleteAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.DeleteAttributeCommand]) (
	*cqrs.Reply[itAttribute.DeleteAttributeResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeSvc.DeleteAttribute)
}

func (this *AttributeHandler) GetAttribute(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.GetAttributeQuery]) (
	*cqrs.Reply[itAttribute.GetAttributeResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeSvc.GetAttribute)
}

func (this *AttributeHandler) SearchAttributes(ctx context.Context, packet *cqrs.RequestPacket[itAttribute.SearchAttributesQuery]) (
	*cqrs.Reply[itAttribute.SearchAttributesResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeSvc.SearchAttributes)
}
