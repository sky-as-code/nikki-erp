package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleHandler(roleSvc it.RoleService, logger logging.LoggerService) *RoleHandler {
	return &RoleHandler{
		Logger:  logger,
		RoleSvc: roleSvc,
	}
}

type RoleHandler struct {
	Logger  logging.LoggerService
	RoleSvc it.RoleService
}

func (this *RoleHandler) CreateRole(ctx context.Context, packet *cqrs.RequestPacket[it.CreateRoleCommand]) (*cqrs.Reply[it.CreateRoleResult], error) {
	cmd := packet.Request()
	result, err := this.RoleSvc.CreateRole(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateRoleResult]{
		Result: *result,
	}
	return reply, nil
}

// func (this *ResourceHandler) UpdateResource(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateResourceCommand]) (*cqrs.Reply[it.UpdateResourceResult], error) {
// 	cmd := packet.Request()
// 	result, err := this.ResourceSvc.UpdateResource(ctx, *cmd)
// 	ft.PanicOnErr(err)

// 	reply := &cqrs.Reply[it.UpdateResourceResult]{
// 		Result: *result,
// 	}
// 	return reply, nil
// }

func (this *RoleHandler) GetRoleById(ctx context.Context, packet *cqrs.RequestPacket[it.GetRoleByIdQuery]) (*cqrs.Reply[it.GetRoleByIdResult], error) {
	query := packet.Request()
	result, err := this.RoleSvc.GetRoleById(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetRoleByIdResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *RoleHandler) SearchRoles(ctx context.Context, packet *cqrs.RequestPacket[it.SearchRolesQuery]) (*cqrs.Reply[it.SearchRolesResult], error) {
	cmd := packet.Request()
	result, err := this.RoleSvc.SearchRoles(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.SearchRolesResult]{
		Result: *result,
	}
	return reply, nil
}
