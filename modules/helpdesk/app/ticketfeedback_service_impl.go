package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketfeedback"
)

func NewTicketFeedbackServiceImpl(repo it.TicketFeedbackRepository, cqrsBus cqrs.CqrsBus) it.TicketFeedbackService {
	return &TicketFeedbackServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketFeedbackServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketFeedbackRepository
}

func (this *TicketFeedbackServiceImpl) CreateTicketFeedback(
	ctx corectx.Context, cmd it.CreateTicketFeedbackCommand,
) (*it.CreateTicketFeedbackResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.TicketFeedback, *domain.TicketFeedback]{Action: "create ticketFeedback", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketFeedbackServiceImpl) DeleteTicketFeedback(
	ctx corectx.Context, cmd it.DeleteTicketFeedbackCommand,
) (*it.DeleteTicketFeedbackResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketFeedback", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketFeedbackServiceImpl) GetTicketFeedback(
	ctx corectx.Context, query it.GetTicketFeedbackQuery,
) (*it.GetTicketFeedbackResult, error) {
	return corecrud.GetOne[domain.TicketFeedback](ctx, corecrud.GetOneParam{Action: "get ticketFeedback", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketFeedbackServiceImpl) TicketFeedbackExists(
	ctx corectx.Context, query it.TicketFeedbackExistsQuery,
) (*it.TicketFeedbackExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketFeedback exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketFeedbackServiceImpl) SearchTicketFeedbacks(
	ctx corectx.Context, query it.SearchTicketFeedbacksQuery,
) (*it.SearchTicketFeedbacksResult, error) {
	return corecrud.Search[domain.TicketFeedback](ctx, corecrud.SearchParam{Action: "search ticketFeedbacks", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketFeedbackServiceImpl) UpdateTicketFeedback(
	ctx corectx.Context, cmd it.UpdateTicketFeedbackCommand,
) (*it.UpdateTicketFeedbackResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.TicketFeedback, *domain.TicketFeedback]{Action: "update ticketFeedback", DbRepoGetter: this.repo, Data: cmd})
}
