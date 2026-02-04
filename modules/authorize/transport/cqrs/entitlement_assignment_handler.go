package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
)

func NewEntitlementAssignmentHandler(entitlementAssignmentSvc it.EntitlementAssignmentService) *EntitlementAssignmentHandler {
	return &EntitlementAssignmentHandler{
		EntitlementAssignmentSvc: entitlementAssignmentSvc,
	}
}

type EntitlementAssignmentHandler struct {
	EntitlementAssignmentSvc it.EntitlementAssignmentService
}

func (this *EntitlementAssignmentHandler) GetAllEntitlementAssignmentBySubject(ctx context.Context, packet *cqrs.RequestPacket[it.GetAllEntitlementAssignmentBySubjectQuery]) (*cqrs.Reply[it.GetAllEntitlementAssignmentBySubjectResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.EntitlementAssignmentSvc.FindAllBySubject)
}
