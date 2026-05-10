package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

func NewTicketActivityDomainServiceImpl(repo it.TicketActivityRepository, cqrsBus cqrs.CqrsBus) it.TicketActivityDomainService {
	return &TicketActivityDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketActivityDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketActivityRepository
}

func (this *TicketActivityDomainServiceImpl) CreateTicketActivity(
	ctx corectx.Context, cmd it.CreateTicketActivityCommand,
) (*it.CreateTicketActivityResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TicketActivity, *models.TicketActivity]{Action: "create ticketActivity", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketActivityDomainServiceImpl) DeleteTicketActivity(
	ctx corectx.Context, cmd it.DeleteTicketActivityCommand,
) (*it.DeleteTicketActivityResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketActivity", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketActivityDomainServiceImpl) GetTicketActivity(
	ctx corectx.Context, query it.GetTicketActivityQuery,
) (*it.GetTicketActivityResult, error) {
	return corecrud.GetOne[models.TicketActivity](ctx, corecrud.GetOneParam{Action: "get ticketActivity", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketActivityDomainServiceImpl) TicketActivityExists(
	ctx corectx.Context, query it.TicketActivityExistsQuery,
) (*it.TicketActivityExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketActivity exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketActivityDomainServiceImpl) SearchTicketActivities(
	ctx corectx.Context, query it.SearchTicketActivitiesQuery,
) (*it.SearchTicketActivitiesResult, error) {
	return corecrud.Search[models.TicketActivity](ctx, corecrud.SearchParam{Action: "search ticketActivitys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketActivityDomainServiceImpl) UpdateTicketActivity(
	ctx corectx.Context, cmd it.UpdateTicketActivityCommand,
) (*it.UpdateTicketActivityResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TicketActivity, *models.TicketActivity]{Action: "update ticketActivity", DbRepoGetter: this.repo, Data: cmd})
}
