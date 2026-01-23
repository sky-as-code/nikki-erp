package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role"
)

func NewRoleHandler(roleSvc it.RoleService) *RoleHandler {
	return &RoleHandler{
		RoleSvc: roleSvc,
	}
}

type RoleHandler struct {
	RoleSvc it.RoleService
}

func (this *RoleHandler) CreateRole(ctx context.Context, packet *cqrs.RequestPacket[it.CreateRoleCommand]) (*cqrs.Reply[it.CreateRoleResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSvc.CreateRole)
}

func (this *RoleHandler) GetRoleById(ctx context.Context, packet *cqrs.RequestPacket[it.GetRoleByIdQuery]) (*cqrs.Reply[it.GetRoleByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSvc.GetRoleById)
}

func (this *RoleHandler) SearchRoles(ctx context.Context, packet *cqrs.RequestPacket[it.SearchRolesQuery]) (*cqrs.Reply[it.SearchRolesResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSvc.SearchRoles)
}

func (this *RoleHandler) GetRolesBySubject(ctx context.Context, packet *cqrs.RequestPacket[it.GetRolesBySubjectQuery]) (*cqrs.Reply[it.GetRolesBySubjectResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.RoleSvc.GetRolesBySubject)
}
