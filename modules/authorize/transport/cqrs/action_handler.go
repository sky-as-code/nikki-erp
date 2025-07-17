package cqrs

import (
	"context"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewActionHandler(actionSvc it.ActionService, logger logging.LoggerService) *ActionHandler {
	return &ActionHandler{
		ActionSvc: actionSvc,
	}
}

type ActionHandler struct {
	ActionSvc it.ActionService
}

func (this *ActionHandler) CreateAction(ctx context.Context, packet *cqrs.RequestPacket[it.CreateActionCommand]) (*cqrs.Reply[it.CreateActionResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ActionSvc.CreateAction)
}

func (this *ActionHandler) UpdateAction(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateActionCommand]) (*cqrs.Reply[it.UpdateActionResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ActionSvc.UpdateAction)
}

func (this *ActionHandler) GetActionById(ctx context.Context, packet *cqrs.RequestPacket[it.GetActionByIdQuery]) (*cqrs.Reply[it.GetActionByIdResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ActionSvc.GetActionById)
}

func (this *ActionHandler) SearchActions(ctx context.Context, packet *cqrs.RequestPacket[it.SearchActionsCommand]) (*cqrs.Reply[it.SearchActionsResult], error) {
	return cqrs.HandlePacket(ctx, packet, this.ActionSvc.SearchActions)
}
