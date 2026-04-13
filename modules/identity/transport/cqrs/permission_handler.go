package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/permission"
)

func NewPermissionHandler(permissionSvc it.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		PermissionSvc: permissionSvc,
	}
}

type PermissionHandler struct {
	PermissionSvc it.PermissionService
}

func (this *PermissionHandler) IsAuthorized(ctx context.Context, packet *cqrs.RequestPacket[it.IsAuthorizedQuery]) (*cqrs.Reply[it.IsAuthorizedResult], error) {
	return cqrs.ServePacket(ctx, string(c.IdentityModuleName), packet, this.PermissionSvc.IsAuthorized)
}
