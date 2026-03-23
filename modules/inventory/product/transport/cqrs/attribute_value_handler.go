package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

func NewAttributeValueHandler(attributeValueSvc itAttributeValue.AttributeValueService) *AttributeValueHandler {
	return &AttributeValueHandler{
		AttributeValueSvc: attributeValueSvc,
	}
}

type AttributeValueHandler struct {
	AttributeValueSvc itAttributeValue.AttributeValueService
}

func (this *AttributeValueHandler) CreateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.CreateAttributeValueCommand]) (*cqrs.Reply[itAttributeValue.CreateAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.CreateAttributeValue)
}

func (this *AttributeValueHandler) UpdateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.UpdateAttributeValueCommand]) (*cqrs.Reply[itAttributeValue.UpdateAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.UpdateAttributeValue)
}

func (this *AttributeValueHandler) DeleteAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.DeleteAttributeValueCommand]) (*cqrs.Reply[itAttributeValue.DeleteAttributeValueResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.DeleteAttributeValue)
}

func (this *AttributeValueHandler) GetAttributeValueById(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.GetAttributeValueByIdQuery]) (*cqrs.Reply[itAttributeValue.GetAttributeValueByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.GetAttributeValueById)
}

func (this *AttributeValueHandler) SearchAttributeValues(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.SearchAttributeValuesQuery]) (*cqrs.Reply[itAttributeValue.SearchAttributeValuesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AttributeValueSvc.SearchAttributeValues)
}
