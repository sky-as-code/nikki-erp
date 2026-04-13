package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

func NewAttributeValueHandler(attributeValueSvc itAttributeValue.AttributeValueService, logger logging.LoggerService) *AttributeValueHandler {
	return &AttributeValueHandler{
		Logger:            logger,
		AttributeValueSvc: attributeValueSvc,
	}
}

type AttributeValueHandler struct {
	Logger            logging.LoggerService
	AttributeValueSvc itAttributeValue.AttributeValueService
}

func (this *AttributeValueHandler) CreateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.CreateAttributeValueCommand]) (
	*cqrs.Reply[itAttributeValue.CreateAttributeValueResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeValueSvc.CreateAttributeValue)
}

func (this *AttributeValueHandler) UpdateAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.UpdateAttributeValueCommand]) (
	*cqrs.Reply[itAttributeValue.UpdateAttributeValueResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeValueSvc.UpdateAttributeValue)
}

func (this *AttributeValueHandler) DeleteAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.DeleteAttributeValueCommand]) (
	*cqrs.Reply[itAttributeValue.DeleteAttributeValueResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeValueSvc.DeleteAttributeValue)
}

func (this *AttributeValueHandler) GetAttributeValue(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.GetAttributeValueQuery]) (
	*cqrs.Reply[itAttributeValue.GetAttributeValueResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeValueSvc.GetAttributeValue)
}

func (this *AttributeValueHandler) SearchAttributeValues(ctx context.Context, packet *cqrs.RequestPacket[itAttributeValue.SearchAttributeValuesQuery]) (
	*cqrs.Reply[itAttributeValue.SearchAttributeValuesResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeValueSvc.SearchAttributeValues)
}
