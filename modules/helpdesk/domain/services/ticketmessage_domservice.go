package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketmessage"
)

func NewTicketMessageDomainServiceImpl(repo it.TicketMessageRepository, cqrsBus cqrs.CqrsBus) it.TicketMessageDomainService {
	return &TicketMessageDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketMessageDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketMessageRepository
}

func (this *TicketMessageDomainServiceImpl) CreateTicketMessage(
	ctx corectx.Context, cmd it.CreateTicketMessageCommand,
) (*it.CreateTicketMessageResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.TicketMessage, *models.TicketMessage]{Action: "create ticketMessage", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketMessageDomainServiceImpl) DeleteTicketMessage(
	ctx corectx.Context, cmd it.DeleteTicketMessageCommand,
) (*it.DeleteTicketMessageResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketMessage", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketMessageDomainServiceImpl) GetTicketMessage(
	ctx corectx.Context, query it.GetTicketMessageQuery,
) (*it.GetTicketMessageResult, error) {
	return corecrud.GetOne[models.TicketMessage](ctx, corecrud.GetOneParam{Action: "get ticketMessage", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketMessageDomainServiceImpl) TicketMessageExists(
	ctx corectx.Context, query it.TicketMessageExistsQuery,
) (*it.TicketMessageExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketMessage exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketMessageDomainServiceImpl) SearchTicketMessages(
	ctx corectx.Context, query it.SearchTicketMessagesQuery,
) (*it.SearchTicketMessagesResult, error) {
	return corecrud.Search[models.TicketMessage](ctx, corecrud.SearchParam{Action: "search ticketMessages", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketMessageDomainServiceImpl) UpdateTicketMessage(
	ctx corectx.Context, cmd it.UpdateTicketMessageCommand,
) (*it.UpdateTicketMessageResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.TicketMessage, *models.TicketMessage]{Action: "update ticketMessage", DbRepoGetter: this.repo, Data: cmd})
}
