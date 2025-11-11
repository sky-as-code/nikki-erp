package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/interfaces"
)

func NewAttributeValueHandler(attributeValueSvc it.AttributeValueService) *AttributeValueHandler {
	return &AttributeValueHandler{
		AttributeValueSvc: attributeValueSvc,
	}
}

type AttributeValueHandler struct {
	AttributeValueSvc it.AttributeValueService
}

func (this *AttributeValueHandler) CreateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[it.CreateAttributeValueCommand]) (*cqrs.Reply[it.CreateAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.CreateAttributeValue)
}

func (this *AttributeValueHandler) UpdateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateAttributeValueCommand]) (*cqrs.Reply[it.UpdateAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.UpdateAttributeValue)
}

func (this *AttributeValueHandler) DeleteAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteAttributeValueCommand]) (*cqrs.Reply[it.DeleteAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.DeleteAttributeValue)
}

func (this *AttributeValueHandler) GetAttributeValueById(ctx context.Context, packet *cqrs.RequestPacket[it.GetAttributeValueByIdQuery]) (*cqrs.Reply[it.GetAttributeValueByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.GetAttributeValueById)
}

func (this *AttributeValueHandler) SearchAttributeValues(ctx context.Context, packet *cqrs.RequestPacket[it.SearchAttributeValuesQuery]) (*cqrs.Reply[it.SearchAttributeValuesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.SearchAttributeValues)
}
