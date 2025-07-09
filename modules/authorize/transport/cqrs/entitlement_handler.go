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
