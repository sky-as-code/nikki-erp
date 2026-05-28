package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

func NewTicketActivityApplicationServiceImpl(ticketActivitySvc it.TicketActivityDomainService) it.TicketActivityAppService {
	return &TicketActivityApplicationServiceImpl{ticketActivitySvc: ticketActivitySvc}
}

type TicketActivityApplicationServiceImpl struct {
	ticketActivitySvc it.TicketActivityDomainService
}

func (this *TicketActivityApplicationServiceImpl) CreateTicketActivity(ctx corectx.Context, cmd it.CreateTicketActivityCommand) (*it.CreateTicketActivityResult, error) {
	return this.ticketActivitySvc.CreateTicketActivity(ctx, cmd)
}

func (this *TicketActivityApplicationServiceImpl) DeleteTicketActivity(ctx corectx.Context, cmd it.DeleteTicketActivityCommand) (*it.DeleteTicketActivityResult, error) {
	return this.ticketActivitySvc.DeleteTicketActivity(ctx, cmd)
}

func (this *TicketActivityApplicationServiceImpl) GetTicketActivity(ctx corectx.Context, query it.GetTicketActivityQuery) (*it.GetTicketActivityResult, error) {
	return this.ticketActivitySvc.GetTicketActivity(ctx, query)
}

func (this *TicketActivityApplicationServiceImpl) TicketActivityExists(ctx corectx.Context, query it.TicketActivityExistsQuery) (*it.TicketActivityExistsResult, error) {
	return this.ticketActivitySvc.TicketActivityExists(ctx, query)
}

func (this *TicketActivityApplicationServiceImpl) SearchTicketActivities(ctx corectx.Context, query it.SearchTicketActivitiesQuery) (*it.SearchTicketActivitiesResult, error) {
	return this.ticketActivitySvc.SearchTicketActivities(ctx, query)
}

func (this *TicketActivityApplicationServiceImpl) UpdateTicketActivity(ctx corectx.Context, cmd it.UpdateTicketActivityCommand) (*it.UpdateTicketActivityResult, error) {
	return this.ticketActivitySvc.UpdateTicketActivity(ctx, cmd)
}
