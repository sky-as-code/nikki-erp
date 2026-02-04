package cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
)

func NewAuthorizeHandler(authorizeSvc it.AuthorizeService, logger logging.LoggerService) *AuthorizeHandler {
	return &AuthorizeHandler{
		AuthorizeSvc: authorizeSvc,
	}
}

type AuthorizeHandler struct {
	AuthorizeSvc it.AuthorizeService
}

func (this *AuthorizeHandler) IsAuthorized(ctx context.Context, packet *cqrs.RequestPacket[it.IsAuthorizedQuery]) (*cqrs.Reply[it.IsAuthorizedResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AuthorizeSvc.IsAuthorized)
}

func (this *AuthorizeHandler) PermissionSnapshot(ctx context.Context, packet *cqrs.RequestPacket[it.PermissionSnapshotQuery]) (*cqrs.Reply[it.PermissionSnapshotResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.AuthorizeSvc.PermissionSnapshot)
}
