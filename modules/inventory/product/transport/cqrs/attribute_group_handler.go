package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

func NewAttributeGroupHandler(attributeGroupSvc itAttributeGroup.AttributeGroupService) *AttributeGroupHandler {
	return &AttributeGroupHandler{
		AttributeGroupSvc: attributeGroupSvc,
	}
}

type AttributeGroupHandler struct {
	AttributeGroupSvc itAttributeGroup.AttributeGroupService
}

func (h *AttributeGroupHandler) CreateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.CreateAttributeGroupCommand]) (*cqrs.Reply[itAttributeGroup.CreateAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.CreateAttributeGroup)
}

func (h *AttributeGroupHandler) UpdateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.UpdateAttributeGroupCommand]) (*cqrs.Reply[itAttributeGroup.UpdateAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.UpdateAttributeGroup)
}

func (h *AttributeGroupHandler) DeleteAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.DeleteAttributeGroupCommand]) (*cqrs.Reply[itAttributeGroup.DeleteAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.DeleteAttributeGroup)
}

func (h *AttributeGroupHandler) GetAttributeGroupById(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.GetAttributeGroupByIdQuery]) (*cqrs.Reply[itAttributeGroup.GetAttributeGroupByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.GetAttributeGroupById)
}

func (h *AttributeGroupHandler) SearchAttributeGroups(ctx context.Context, packet *cqrs.RequestPacket[itAttributeGroup.SearchAttributeGroupsQuery]) (*cqrs.Reply[itAttributeGroup.SearchAttributeGroupsResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.SearchAttributeGroups)
}
