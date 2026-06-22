package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

func NewPermissionHandler(permissionSvc it.PermissionAppService) *PermissionHandler {
	return &PermissionHandler{
		permissionSvc: permissionSvc,
	}
}

type PermissionHandler struct {
	permissionSvc it.PermissionAppService
}

func (this *PermissionHandler) IsAuthorized(ctx context.Context, packet *cqrs.RequestPacket[it.IsAuthorizedQuery]) (*cqrs.Reply[it.IsAuthorizedResult], error) {
	return cqrs.ServePacket(ctx, string(c.IamModuleName), packet, this.permissionSvc.IsAuthorized)
}

func (this *PermissionHandler) GetUserEntitlements(ctx context.Context, packet *cqrs.RequestPacket[it.GetUserEntitlementsQuery]) (*cqrs.Reply[it.GetUserEntitlementsResult], error) {
	return cqrs.ServePacket(ctx, string(c.IamModuleName), packet, this.permissionSvc.GetUserEntitlements)
}
