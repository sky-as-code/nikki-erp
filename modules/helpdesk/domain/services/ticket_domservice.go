package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticket"
)

func NewTicketDomainServiceImpl(repo it.TicketRepository, cqrsBus cqrs.CqrsBus) it.TicketDomainService {
	return &TicketDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type TicketDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.TicketRepository
}

func (this *TicketDomainServiceImpl) CreateTicket(
	ctx corectx.Context, cmd it.CreateTicketCommand,
) (*it.CreateTicketResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Ticket, *models.Ticket]{Action: "create ticket", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *TicketDomainServiceImpl) DeleteTicket(
	ctx corectx.Context, cmd it.DeleteTicketCommand,
) (*it.DeleteTicketResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete ticket", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *TicketDomainServiceImpl) GetTicket(
	ctx corectx.Context, query it.GetTicketQuery,
) (*it.GetTicketResult, error) {
	return corecrud.GetOne[models.Ticket](ctx, corecrud.GetOneParam{Action: "get ticket", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *TicketDomainServiceImpl) TicketExists(
	ctx corectx.Context, query it.TicketExistsQuery,
) (*it.TicketExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if ticket exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *TicketDomainServiceImpl) SearchTickets(
	ctx corectx.Context, query it.SearchTicketsQuery,
) (*it.SearchTicketsResult, error) {
	return corecrud.Search[models.Ticket](ctx, corecrud.SearchParam{Action: "search tickets", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *TicketDomainServiceImpl) UpdateTicket(
	ctx corectx.Context, cmd it.UpdateTicketCommand,
) (*it.UpdateTicketResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Ticket, *models.Ticket]{Action: "update ticket", DbRepoGetter: this.repo, Data: cmd})
}

func (this *TicketDomainServiceImpl) SetTicketIsArchived(
	ctx corectx.Context, cmd it.SetTicketIsArchivedCommand,
) (*it.SetTicketIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}

func (this *TicketDomainServiceImpl) ManageTicketCategories(
	ctx corectx.Context, cmd it.ManageTicketCategoriesCommand,
) (*it.ManageTicketCategoriesResult, error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage ticket categories",
		DbRepoGetter:       this.repo,
		DestSchemaName:     models.TicketCategorySchemaName,
		SrcId:              cmd.TicketId,
		SrcIdFieldForError: "ticket_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
		BeforeInsert: func(_ corectx.Context, dbRecords []dmodel.DynamicFields) error {
			ulidType := dmodel.FieldDataTypeUlid()
			for _, rec := range dbRecords {
				rec[basemodel.FieldId] = *ulidType.DefaultValue().Get()
			}
			return nil
		},
	})
}
