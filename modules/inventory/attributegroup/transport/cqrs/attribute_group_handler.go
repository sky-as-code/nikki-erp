package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup/interfaces"
)

func NewAttributeGroupHandler(attributeGroupSvc it.AttributeGroupService) *AttributeGroupHandler {
	return &AttributeGroupHandler{
		AttributeGroupSvc: attributeGroupSvc,
	}
}

type AttributeGroupHandler struct {
	AttributeGroupSvc it.AttributeGroupService
}

func (h *AttributeGroupHandler) CreateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[it.CreateAttributeGroupCommand]) (*cqrs.Reply[it.CreateAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.CreateAttributeGroup)
}

func (h *AttributeGroupHandler) UpdateAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateAttributeGroupCommand]) (*cqrs.Reply[it.UpdateAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.UpdateAttributeGroup)
}

func (h *AttributeGroupHandler) DeleteAttributeGroup(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteAttributeGroupCommand]) (*cqrs.Reply[it.DeleteAttributeGroupResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.DeleteAttributeGroup)
}

func (h *AttributeGroupHandler) GetAttributeGroupById(ctx context.Context, packet *cqrs.RequestPacket[it.GetAttributeGroupByIdQuery]) (*cqrs.Reply[it.GetAttributeGroupByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.GetAttributeGroupById)
}

func (h *AttributeGroupHandler) SearchAttributeGroups(ctx context.Context, packet *cqrs.RequestPacket[it.SearchAttributeGroupsQuery]) (*cqrs.Reply[it.SearchAttributeGroupsResult], error) {
	return cqrs.HandlePacket(ctx, packet, h.AttributeGroupSvc.SearchAttributeGroups)
}
