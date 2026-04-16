package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

func NewTicketActivityServiceImpl(repo it.TicketActivityRepository, cqrsBus cqrs.CqrsBus) it.TicketActivityService {
	return &TicketActivityServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketActivityServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketActivityRepository
}

func (this *TicketActivityServiceImpl) CreateTicketActivity(
	ctx corectx.Context, cmd it.CreateTicketActivityCommand,
) (*it.CreateTicketActivityResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.TicketActivity, *domain.TicketActivity]{Action: "create ticketActivity", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketActivityServiceImpl) DeleteTicketActivity(
	ctx corectx.Context, cmd it.DeleteTicketActivityCommand,
) (*it.DeleteTicketActivityResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketActivity", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketActivityServiceImpl) GetTicketActivity(
	ctx corectx.Context, query it.GetTicketActivityQuery,
) (*it.GetTicketActivityResult, error) {
	return corecrud.GetOne[domain.TicketActivity](ctx, corecrud.GetOneParam{Action: "get ticketActivity", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketActivityServiceImpl) TicketActivityExists(
	ctx corectx.Context, query it.TicketActivityExistsQuery,
) (*it.TicketActivityExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketActivity exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketActivityServiceImpl) SearchTicketActivities(
	ctx corectx.Context, query it.SearchTicketActivitiesQuery,
) (*it.SearchTicketActivitiesResult, error) {
	return corecrud.Search[domain.TicketActivity](ctx, corecrud.SearchParam{Action: "search ticketActivitys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketActivityServiceImpl) UpdateTicketActivity(
	ctx corectx.Context, cmd it.UpdateTicketActivityCommand,
) (*it.UpdateTicketActivityResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.TicketActivity, *domain.TicketActivity]{Action: "update ticketActivity", DbRepoGetter: this.repo, Data: cmd})
}
