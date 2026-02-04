package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
)

func NewEntitlementHandler(entitlementSvc it.EntitlementService) *EntitlementHandler {
	return &EntitlementHandler{
		EntitlementSvc: entitlementSvc,
	}
}

type EntitlementHandler struct {
	EntitlementSvc it.EntitlementService
}

func (this *EntitlementHandler) CreateEntitlement(ctx context.Context, packet *cqrs.RequestPacket[it.CreateEntitlementCommand]) (*cqrs.Reply[it.CreateEntitlementResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.CreateEntitlement)
}

func (this *EntitlementHandler) EntitlementExists(ctx context.Context, packet *cqrs.RequestPacket[it.EntitlementExistsQuery]) (*cqrs.Reply[it.EntitlementExistsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.EntitlementExists)
}

func (this *EntitlementHandler) UpdateEntitlement(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateEntitlementCommand]) (*cqrs.Reply[it.UpdateEntitlementResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.UpdateEntitlement)
}

func (this *EntitlementHandler) GetEntitlementById(ctx context.Context, packet *cqrs.RequestPacket[it.GetEntitlementByIdQuery]) (*cqrs.Reply[it.GetEntitlementByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.GetEntitlementById)
}

func (this *EntitlementHandler) GetAllEntitlementByIds(ctx context.Context, packet *cqrs.RequestPacket[it.GetAllEntitlementByIdsQuery]) (*cqrs.Reply[it.GetAllEntitlementByIdsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.GetAllEntitlementByIds)
}

func (this *EntitlementHandler) SearchEntitlements(ctx context.Context, packet *cqrs.RequestPacket[it.SearchEntitlementsQuery]) (*cqrs.Reply[it.SearchEntitlementsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementSvc.SearchEntitlements)
}
