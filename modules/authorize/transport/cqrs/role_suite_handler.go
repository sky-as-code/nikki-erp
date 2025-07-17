package cqrs

import (
	"context"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func NewRoleSuiteHandler(roleSuiteSvc it.RoleSuiteService) *RoleSuiteHandler {
	return &RoleSuiteHandler{
		RoleSuiteSvc: roleSuiteSvc,
	}
}

type RoleSuiteHandler struct {
	RoleSuiteSvc it.RoleSuiteService
}

func (this *RoleSuiteHandler) GetRoleSuitesBySubject(ctx context.Context, packet *cqrs.RequestPacket[it.GetRoleSuitesBySubjectQuery]) (*cqrs.Reply[it.GetRoleSuitesBySubjectResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSuiteSvc.GetRoleSuitesBySubject)
}

func (this *RoleSuiteHandler) GetRoleSuiteById(ctx context.Context, packet *cqrs.RequestPacket[it.GetRoleSuiteByIdQuery]) (*cqrs.Reply[it.GetRoleSuiteByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSuiteSvc.GetRoleSuiteById)
}
