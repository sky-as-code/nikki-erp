package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewActionHandler(actionSvc it.ActionService, logger logging.LoggerService) *ActionHandler {
	return &ActionHandler{
		Logger:    logger,
		ActionSvc: actionSvc,
	}
}

type ActionHandler struct {
	Logger    logging.LoggerService
	ActionSvc it.ActionService
}

func (this *ActionHandler) CreateAction(ctx context.Context, packet *cqrs.RequestPacket[it.CreateActionCommand]) (*cqrs.Reply[it.CreateActionResult], error) {
	cmd := packet.Request()
	result, err := this.ActionSvc.CreateAction(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.CreateActionResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ActionHandler) UpdateAction(ctx context.Context, packet *cqrs.RequestPacket[it.UpdateActionCommand]) (*cqrs.Reply[it.UpdateActionResult], error) {
	cmd := packet.Request()
	result, err := this.ActionSvc.UpdateAction(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.UpdateActionResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ActionHandler) GetActionById(ctx context.Context, packet *cqrs.RequestPacket[it.GetActionByIdQuery]) (*cqrs.Reply[it.GetActionByIdResult], error) {
	query := packet.Request()
	result, err := this.ActionSvc.GetActionById(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetActionByIdResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *ActionHandler) SearchActions(ctx context.Context, packet *cqrs.RequestPacket[it.SearchActionsCommand]) (*cqrs.Reply[it.SearchActionsResult], error) {
	cmd := packet.Request()
	result, err := this.ActionSvc.SearchActions(ctx, *cmd)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.SearchActionsResult]{
		Result: *result,
	}
	return reply, nil
}
