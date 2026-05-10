package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketassignment"
)

func NewTicketAssignmentApplicationServiceImpl(ticketAssignmentSvc it.TicketAssignmentDomainService) it.TicketAssignmentAppService {
	return &TicketAssignmentApplicationServiceImpl{ticketAssignmentSvc: ticketAssignmentSvc}
}

type TicketAssignmentApplicationServiceImpl struct {
	ticketAssignmentSvc it.TicketAssignmentDomainService
}

func (this *TicketAssignmentApplicationServiceImpl) CreateTicketAssignment(ctx corectx.Context, cmd it.CreateTicketAssignmentCommand) (*it.CreateTicketAssignmentResult, error) {
	return this.ticketAssignmentSvc.CreateTicketAssignment(ctx, cmd)
}

func (this *TicketAssignmentApplicationServiceImpl) DeleteTicketAssignment(ctx corectx.Context, cmd it.DeleteTicketAssignmentCommand) (*it.DeleteTicketAssignmentResult, error) {
	return this.ticketAssignmentSvc.DeleteTicketAssignment(ctx, cmd)
}

func (this *TicketAssignmentApplicationServiceImpl) GetTicketAssignment(ctx corectx.Context, query it.GetTicketAssignmentQuery) (*it.GetTicketAssignmentResult, error) {
	return this.ticketAssignmentSvc.GetTicketAssignment(ctx, query)
}

func (this *TicketAssignmentApplicationServiceImpl) TicketAssignmentExists(ctx corectx.Context, query it.TicketAssignmentExistsQuery) (*it.TicketAssignmentExistsResult, error) {
	return this.ticketAssignmentSvc.TicketAssignmentExists(ctx, query)
}

func (this *TicketAssignmentApplicationServiceImpl) SearchTicketAssignments(ctx corectx.Context, query it.SearchTicketAssignmentsQuery) (*it.SearchTicketAssignmentsResult, error) {
	return this.ticketAssignmentSvc.SearchTicketAssignments(ctx, query)
}

func (this *TicketAssignmentApplicationServiceImpl) UpdateTicketAssignment(ctx corectx.Context, cmd it.UpdateTicketAssignmentCommand) (*it.UpdateTicketAssignmentResult, error) {
	return this.ticketAssignmentSvc.UpdateTicketAssignment(ctx, cmd)
}
