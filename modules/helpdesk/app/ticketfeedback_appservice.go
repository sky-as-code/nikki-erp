package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketfeedback"
)

func NewTicketFeedbackApplicationServiceImpl(ticketFeedbackSvc it.TicketFeedbackDomainService) it.TicketFeedbackAppService {
	return &TicketFeedbackApplicationServiceImpl{ticketFeedbackSvc: ticketFeedbackSvc}
}

type TicketFeedbackApplicationServiceImpl struct {
	ticketFeedbackSvc it.TicketFeedbackDomainService
}

func (this *TicketFeedbackApplicationServiceImpl) CreateTicketFeedback(ctx corectx.Context, cmd it.CreateTicketFeedbackCommand) (*it.CreateTicketFeedbackResult, error) {
	return this.ticketFeedbackSvc.CreateTicketFeedback(ctx, cmd)
}

func (this *TicketFeedbackApplicationServiceImpl) DeleteTicketFeedback(ctx corectx.Context, cmd it.DeleteTicketFeedbackCommand) (*it.DeleteTicketFeedbackResult, error) {
	return this.ticketFeedbackSvc.DeleteTicketFeedback(ctx, cmd)
}

func (this *TicketFeedbackApplicationServiceImpl) GetTicketFeedback(ctx corectx.Context, query it.GetTicketFeedbackQuery) (*it.GetTicketFeedbackResult, error) {
	return this.ticketFeedbackSvc.GetTicketFeedback(ctx, query)
}

func (this *TicketFeedbackApplicationServiceImpl) TicketFeedbackExists(ctx corectx.Context, query it.TicketFeedbackExistsQuery) (*it.TicketFeedbackExistsResult, error) {
	return this.ticketFeedbackSvc.TicketFeedbackExists(ctx, query)
}

func (this *TicketFeedbackApplicationServiceImpl) SearchTicketFeedbacks(ctx corectx.Context, query it.SearchTicketFeedbacksQuery) (*it.SearchTicketFeedbacksResult, error) {
	return this.ticketFeedbackSvc.SearchTicketFeedbacks(ctx, query)
}

func (this *TicketFeedbackApplicationServiceImpl) UpdateTicketFeedback(ctx corectx.Context, cmd it.UpdateTicketFeedbackCommand) (*it.UpdateTicketFeedbackResult, error) {
	return this.ticketFeedbackSvc.UpdateTicketFeedback(ctx, cmd)
}
