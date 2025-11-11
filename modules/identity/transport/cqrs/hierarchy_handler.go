package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

func NewHierarchyHandler(hierarchySvc it.HierarchyService, logger logging.LoggerService) *HierarchyHandler {
	return &HierarchyHandler{
		Logger:       logger,
		HierarchySvc: hierarchySvc,
	}
}

type HierarchyHandler struct {
	Logger       logging.LoggerService
	HierarchySvc it.HierarchyService
}

func (this *HierarchyHandler) CreateHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.CreateHierarchyLevelCommand]) (*cqrs.Reply[it.CreateHierarchyLevelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.CreateHierarchyLevel)
}

func (this *HierarchyHandler) UpdateHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateHierarchyLevelCommand]) (*cqrs.Reply[it.UpdateHierarchyLevelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.UpdateHierarchyLevel)
}

func (this *HierarchyHandler) DeleteHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteHierarchyLevelCommand]) (*cqrs.Reply[it.DeleteHierarchyLevelResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.DeleteHierarchyLevel)
}

func (this *HierarchyHandler) GetHierarchyLevelById(ctx context.Context, packet *cqrs.RequestPacket[it.GetHierarchyLevelByIdQuery]) (*cqrs.Reply[it.GetHierarchyLevelByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.GetHierarchyLevelById)
}

func (this *HierarchyHandler) SearchHierarchyLevels(ctx context.Context, packet *cqrs.RequestPacket[it.SearchHierarchyLevelsQuery]) (*cqrs.Reply[it.SearchHierarchyLevelsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.SearchHierarchyLevels)
}

func (this *HierarchyHandler) ExistsHierarchyById(ctx context.Context, packet *cqrs.RequestPacket[it.ExistsHierarchyLevelByIdQuery]) (*cqrs.Reply[it.ExistsHierarchyLevelByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.HierarchySvc.ExistsHierarchyById)
}
