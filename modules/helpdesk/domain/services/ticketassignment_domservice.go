package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketassignment"
)

func NewTicketAssignmentDomainServiceImpl(repo it.TicketAssignmentRepository, cqrsBus cqrs.CqrsBus) it.TicketAssignmentDomainService {
	return &TicketAssignmentDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketAssignmentDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketAssignmentRepository
}

func (this *TicketAssignmentDomainServiceImpl) CreateTicketAssignment(
	ctx corectx.Context, cmd it.CreateTicketAssignmentCommand,
) (*it.CreateTicketAssignmentResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TicketAssignment, *models.TicketAssignment]{Action: "create ticketAssignment", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketAssignmentDomainServiceImpl) DeleteTicketAssignment(
	ctx corectx.Context, cmd it.DeleteTicketAssignmentCommand,
) (*it.DeleteTicketAssignmentResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketAssignment", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketAssignmentDomainServiceImpl) GetTicketAssignment(
	ctx corectx.Context, query it.GetTicketAssignmentQuery,
) (*it.GetTicketAssignmentResult, error) {
	return corecrud.GetOne[models.TicketAssignment](ctx, corecrud.GetOneParam{Action: "get ticketAssignment", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketAssignmentDomainServiceImpl) TicketAssignmentExists(
	ctx corectx.Context, query it.TicketAssignmentExistsQuery,
) (*it.TicketAssignmentExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketAssignment exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketAssignmentDomainServiceImpl) SearchTicketAssignments(
	ctx corectx.Context, query it.SearchTicketAssignmentsQuery,
) (*it.SearchTicketAssignmentsResult, error) {
	return corecrud.Search[models.TicketAssignment](ctx, corecrud.SearchParam{Action: "search ticketAssignments", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketAssignmentDomainServiceImpl) UpdateTicketAssignment(
	ctx corectx.Context, cmd it.UpdateTicketAssignmentCommand,
) (*it.UpdateTicketAssignmentResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TicketAssignment, *models.TicketAssignment]{Action: "update ticketAssignment", DbRepoGetter: this.repo, Data: cmd})
}
