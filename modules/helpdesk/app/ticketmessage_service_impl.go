package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketmessage"
)

func NewTicketMessageServiceImpl(repo it.TicketMessageRepository, cqrsBus cqrs.CqrsBus) it.TicketMessageService {
	return &TicketMessageServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketMessageServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketMessageRepository
}

func (this *TicketMessageServiceImpl) CreateTicketMessage(
	ctx corectx.Context, cmd it.CreateTicketMessageCommand,
) (*it.CreateTicketMessageResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.TicketMessage, *domain.TicketMessage]{Action: "create ticketMessage", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketMessageServiceImpl) DeleteTicketMessage(
	ctx corectx.Context, cmd it.DeleteTicketMessageCommand,
) (*it.DeleteTicketMessageResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticketMessage", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketMessageServiceImpl) GetTicketMessage(
	ctx corectx.Context, query it.GetTicketMessageQuery,
) (*it.GetTicketMessageResult, error) {
	return corecrud.GetOne[domain.TicketMessage](ctx, corecrud.GetOneParam{Action: "get ticketMessage", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketMessageServiceImpl) TicketMessageExists(
	ctx corectx.Context, query it.TicketMessageExistsQuery,
) (*it.TicketMessageExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticketMessage exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketMessageServiceImpl) SearchTicketMessages(
	ctx corectx.Context, query it.SearchTicketMessagesQuery,
) (*it.SearchTicketMessagesResult, error) {
	return corecrud.Search[domain.TicketMessage](ctx, corecrud.SearchParam{Action: "search ticketMessages", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketMessageServiceImpl) UpdateTicketMessage(
	ctx corectx.Context, cmd it.UpdateTicketMessageCommand,
) (*it.UpdateTicketMessageResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.TicketMessage, *domain.TicketMessage]{Action: "update ticketMessage", DbRepoGetter: this.repo, Data: cmd})
}
