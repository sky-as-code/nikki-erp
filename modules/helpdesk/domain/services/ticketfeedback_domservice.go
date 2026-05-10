package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketfeedback"
)

func NewTicketFeedbackDomainServiceImpl(repo it.TicketFeedbackRepository, cqrsBus cqrs.CqrsBus) it.TicketFeedbackDomainService {
	return &TicketFeedbackDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketFeedbackDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketFeedbackRepository
}

func (this *TicketFeedbackDomainServiceImpl) CreateTicketFeedback(
	ctx corectx.Context, cmd it.CreateTicketFeedbackCommand,
) (*it.CreateTicketFeedbackResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TicketFeedback, *models.TicketFeedback]{Action: "create ticketFeedback", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketFeedbackDomainServiceImpl) DeleteTicketFeedback(
	ctx corectx.Context, cmd it.DeleteTicketFeedbackCommand,
) (*it.DeleteTicketFeedbackResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketFeedback", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketFeedbackDomainServiceImpl) GetTicketFeedback(
	ctx corectx.Context, query it.GetTicketFeedbackQuery,
) (*it.GetTicketFeedbackResult, error) {
	return corecrud.GetOne[models.TicketFeedback](ctx, corecrud.GetOneParam{Action: "get ticketFeedback", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketFeedbackDomainServiceImpl) TicketFeedbackExists(
	ctx corectx.Context, query it.TicketFeedbackExistsQuery,
) (*it.TicketFeedbackExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketFeedback exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketFeedbackDomainServiceImpl) SearchTicketFeedbacks(
	ctx corectx.Context, query it.SearchTicketFeedbacksQuery,
) (*it.SearchTicketFeedbacksResult, error) {
	return corecrud.Search[models.TicketFeedback](ctx, corecrud.SearchParam{Action: "search ticketFeedbacks", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketFeedbackDomainServiceImpl) UpdateTicketFeedback(
	ctx corectx.Context, cmd it.UpdateTicketFeedbackCommand,
) (*it.UpdateTicketFeedbackResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TicketFeedback, *models.TicketFeedback]{Action: "update ticketFeedback", DbRepoGetter: this.repo, Data: cmd})
}
