package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketmessage"
)

func NewTicketMessageApplicationServiceImpl(ticketMessageSvc it.TicketMessageDomainService) it.TicketMessageAppService {
	return &TicketMessageApplicationServiceImpl{ticketMessageSvc: ticketMessageSvc}
}

type TicketMessageApplicationServiceImpl struct {
	ticketMessageSvc it.TicketMessageDomainService
}

func (this *TicketMessageApplicationServiceImpl) CreateTicketMessage(ctx corectx.Context, cmd it.CreateTicketMessageCommand) (*it.CreateTicketMessageResult, error) {
	return this.ticketMessageSvc.CreateTicketMessage(ctx, cmd)
}

func (this *TicketMessageApplicationServiceImpl) DeleteTicketMessage(ctx corectx.Context, cmd it.DeleteTicketMessageCommand) (*it.DeleteTicketMessageResult, error) {
	return this.ticketMessageSvc.DeleteTicketMessage(ctx, cmd)
}

func (this *TicketMessageApplicationServiceImpl) GetTicketMessage(ctx corectx.Context, query it.GetTicketMessageQuery) (*it.GetTicketMessageResult, error) {
	return this.ticketMessageSvc.GetTicketMessage(ctx, query)
}

func (this *TicketMessageApplicationServiceImpl) TicketMessageExists(ctx corectx.Context, query it.TicketMessageExistsQuery) (*it.TicketMessageExistsResult, error) {
	return this.ticketMessageSvc.TicketMessageExists(ctx, query)
}

func (this *TicketMessageApplicationServiceImpl) SearchTicketMessages(ctx corectx.Context, query it.SearchTicketMessagesQuery) (*it.SearchTicketMessagesResult, error) {
	return this.ticketMessageSvc.SearchTicketMessages(ctx, query)
}

func (this *TicketMessageApplicationServiceImpl) UpdateTicketMessage(ctx corectx.Context, cmd it.UpdateTicketMessageCommand) (*it.UpdateTicketMessageResult, error) {
	return this.ticketMessageSvc.UpdateTicketMessage(ctx, cmd)
}
