package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
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

func (this *HierarchyHandler) CreateHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.CreateHierarchyLevelCommand]) (
	*cqrs.Reply[it.CreateHierarchyLevelResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.CreateHierarchyLevel)
}

func (this *HierarchyHandler) UpdateHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateHierarchyLevelCommand]) (
	*cqrs.Reply[it.UpdateHierarchyLevelResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.UpdateHierarchyLevel)
}

func (this *HierarchyHandler) DeleteHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.DeleteHierarchyLevelCommand]) (
	*cqrs.Reply[it.DeleteHierarchyLevelResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.DeleteHierarchyLevel)
}

func (this *HierarchyHandler) GetHierarchyLevel(ctx context.Context, packet *cqrs.RequestPacket[it.GetHierarchyLevelQuery]) (
	*cqrs.Reply[it.GetHierarchyLevelResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.GetHierarchyLevel)
}

func (this *HierarchyHandler) ManageHierarchyLevelUsers(ctx context.Context, packet *cqrs.RequestPacket[it.ManageHierarchyLevelUsersCommand]) (
	*cqrs.Reply[it.ManageHierarchyLevelUsersResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.ManageHierarchyLevelUsers)
}

func (this *HierarchyHandler) SearchHierarchyLevels(ctx context.Context, packet *cqrs.RequestPacket[it.SearchHierarchyLevelsQuery]) (
	*cqrs.Reply[it.SearchHierarchyLevelsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.SearchHierarchyLevels)
}

func (this *HierarchyHandler) HierarchyLevelExists(ctx context.Context, packet *cqrs.RequestPacket[it.HierarchyLevelExistsQuery]) (
	*cqrs.Reply[it.HierarchyLevelExistsResult], error,
) {
	return cqrs.HandlePacket2(ctx, string(c.IdentityModuleName), packet, this.HierarchySvc.HierarchyLevelExists)
}
