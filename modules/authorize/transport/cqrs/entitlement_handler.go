package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewEntitlementHandler(entitlementSvc it.EntitlementService, logger logging.LoggerService) *EntitlementHandler {
	return &EntitlementHandler{
		Logger:         logger,
		EntitlementSvc: entitlementSvc,
	}
}

type EntitlementHandler struct {
	Logger         logging.LoggerService
	EntitlementSvc it.EntitlementService
}

func (this *EntitlementHandler) CreateEntitlement(ctx context.Context, packet *cqrs.RequestPacket[it.CreateEntitlementCommand]) (*cqrs.Reply[it.CreateEntitlementResult], error) {
	cmd := packet.Request()
	result, err := this.EntitlementSvc.CreateEntitlement(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateEntitlementResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *EntitlementHandler) EntitlementExists(ctx context.Context, packet *cqrs.RequestPacket[it.EntitlementExistsCommand]) (*cqrs.Reply[it.EntitlementExistsResult], error) {
	cmd := packet.Request()
	result, err := this.EntitlementSvc.EntitlementExists(ctx, *cmd)
	ft.PanicOnErr(err)

	return &cqrs.Reply[it.EntitlementExistsResult]{
		Result: *result,
	}, nil
}

func (this *EntitlementHandler) UpdateEntitlement(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateEntitlementCommand]) (*cqrs.Reply[it.UpdateEntitlementResult], error) {
	cmd := packet.Request()
	result, err := this.EntitlementSvc.UpdateEntitlement(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateEntitlementResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *EntitlementHandler) GetEntitlementById(ctx context.Context, packet *cqrs.RequestPacket[it.GetEntitlementByIdQuery]) (*cqrs.Reply[it.GetEntitlementByIdResult], error) {
	query := packet.Request()
	result, err := this.EntitlementSvc.GetEntitlementById(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetEntitlementByIdResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *EntitlementHandler) GetAllEntitlementByIds(ctx context.Context, packet *cqrs.RequestPacket[it.GetAllEntitlementByIdsQuery]) (*cqrs.Reply[it.GetAllEntitlementByIdsResult], error) {
	query := packet.Request()
	result, err := this.EntitlementSvc.GetAllEntitlementByIds(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetAllEntitlementByIdsResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *EntitlementHandler) SearchEntitlements(ctx context.Context, packet *cqrs.RequestPacket[it.SearchEntitlementsQuery]) (*cqrs.Reply[it.SearchEntitlementsResult], error) {
	query := packet.Request()
	result, err := this.EntitlementSvc.SearchEntitlements(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.SearchEntitlementsResult]{
		Result: *result,
	}
	return reply, nil
}
