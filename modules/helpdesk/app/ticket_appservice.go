package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticket"
)

func NewTicketApplicationServiceImpl(ticketSvc it.TicketDomainService) it.TicketAppService {
	return &TicketApplicationServiceImpl{ticketSvc: ticketSvc}
}

type TicketApplicationServiceImpl struct {
	ticketSvc it.TicketDomainService
}

func (this *TicketApplicationServiceImpl) CreateTicket(ctx corectx.Context, cmd it.CreateTicketCommand) (*it.CreateTicketResult, error) {
	return this.ticketSvc.CreateTicket(ctx, cmd)
}

func (this *TicketApplicationServiceImpl) DeleteTicket(ctx corectx.Context, cmd it.DeleteTicketCommand) (*it.DeleteTicketResult, error) {
	return this.ticketSvc.DeleteTicket(ctx, cmd)
}

func (this *TicketApplicationServiceImpl) GetTicket(ctx corectx.Context, query it.GetTicketQuery) (*it.GetTicketResult, error) {
	return this.ticketSvc.GetTicket(ctx, query)
}

func (this *TicketApplicationServiceImpl) TicketExists(ctx corectx.Context, query it.TicketExistsQuery) (*it.TicketExistsResult, error) {
	return this.ticketSvc.TicketExists(ctx, query)
}

func (this *TicketApplicationServiceImpl) SearchTickets(ctx corectx.Context, query it.SearchTicketsQuery) (*it.SearchTicketsResult, error) {
	return this.ticketSvc.SearchTickets(ctx, query)
}

func (this *TicketApplicationServiceImpl) UpdateTicket(ctx corectx.Context, cmd it.UpdateTicketCommand) (*it.UpdateTicketResult, error) {
	return this.ticketSvc.UpdateTicket(ctx, cmd)
}

func (this *TicketApplicationServiceImpl) SetTicketIsArchived(ctx corectx.Context, cmd it.SetTicketIsArchivedCommand) (*it.SetTicketIsArchivedResult, error) {
	return this.ticketSvc.SetTicketIsArchived(ctx, cmd)
}

func (this *TicketApplicationServiceImpl) ManageTicketCategories(ctx corectx.Context, cmd it.ManageTicketCategoriesCommand) (*it.ManageTicketCategoriesResult, error) {
	return this.ticketSvc.ManageTicketCategories(ctx, cmd)
}
