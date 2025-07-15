package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewEntitlementAssignmentHandler(entitlementAssignmentSvc it.EntitlementAssignmentService, logger logging.LoggerService) *EntitlementAssignmentHandler {
	return &EntitlementAssignmentHandler{
		Logger:    logger,
		EntitlementAssignmentSvc: entitlementAssignmentSvc,
	}
}

type EntitlementAssignmentHandler struct {
	Logger    logging.LoggerService
	EntitlementAssignmentSvc it.EntitlementAssignmentService
}

func (this *EntitlementAssignmentHandler) GetAllEntitlementAssignmentBySubject(ctx context.Context, packet *cqrs.RequestPacket[it.GetAllEntitlementAssignmentBySubjectQuery]) (*cqrs.Reply[it.GetAllEntitlementAssignmentBySubjectResult], error) {
	query := packet.Request()
	result, err := this.EntitlementAssignmentSvc.FindAllBySubject(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetAllEntitlementAssignmentBySubjectResult]{
		Result: *result,
	}
	return reply, nil
}
