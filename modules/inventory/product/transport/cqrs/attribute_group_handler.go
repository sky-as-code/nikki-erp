package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

func NewAttributeGroupHandler(attributeGroupSvc itAttributeGroup.AttributeGroupService, logger logging.LoggerService) *AttributeGroupHandler {
	return &AttributeGroupHandler{
		Logger:            logger,
		AttributeGroupSvc: attributeGroupSvc,
	}
}

type AttributeGroupHandler struct {
	Logger            logging.LoggerService
	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func (this *AttributeGroupHandler) CreateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.CreateAttributeGroupCommand]) (
	*cqrs.Reply[itAttributeGroup.CreateAttributeGroupResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeGroupSvc.CreateAttributeGroup)
}

func (this *AttributeGroupHandler) UpdateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.UpdateAttributeGroupCommand]) (
	*cqrs.Reply[itAttributeGroup.UpdateAttributeGroupResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeGroupSvc.UpdateAttributeGroup)
}

func (this *AttributeGroupHandler) DeleteAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.DeleteAttributeGroupCommand]) (
	*cqrs.Reply[itAttributeGroup.DeleteAttributeGroupResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeGroupSvc.DeleteAttributeGroup)
}

func (this *AttributeGroupHandler) GetAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.GetAttributeGroupQuery]) (
	*cqrs.Reply[itAttributeGroup.GetAttributeGroupResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeGroupSvc.GetAttributeGroup)
}

func (this *AttributeGroupHandler) SearchAttributeGroups(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.SearchAttributeGroupsQuery]) (
	*cqrs.Reply[itAttributeGroup.SearchAttributeGroupsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.InventoryModuleName), packet, this.AttributeGroupSvc.SearchAttributeGroups)
}
