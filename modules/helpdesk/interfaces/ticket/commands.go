package ticket

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketCommand)(nil)
	req = (*DeleteTicketCommand)(nil)
	req = (*GetTicketQuery)(nil)
	req = (*TicketExistsQuery)(nil)
	req = (*SearchTicketsQuery)(nil)
	req = (*UpdateTicketCommand)(nil)
	req = (*SetTicketIsArchivedCommand)(nil)
	req = (*ManageTicketCategoriesCommand)(nil)
	util.Unused(req)
}

var createTicketCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "createTicket"}

type CreateTicketCommand struct{ domain.Ticket }

func (CreateTicketCommand) CqrsRequestType() cqrs.RequestType { return createTicketCommandType }
func (CreateTicketCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketSchemaName)
}

type CreateTicketResult = dyn.OpResult[domain.Ticket]

var deleteTicketCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "deleteTicket"}

type DeleteTicketCommand dyn.DeleteOneCommand

func (DeleteTicketCommand) CqrsRequestType() cqrs.RequestType { return deleteTicketCommandType }

type DeleteTicketResult = dyn.OpResult[dyn.MutateResultData]

var getTicketQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "getTicket"}

type GetTicketQuery dyn.GetOneQuery

func (GetTicketQuery) CqrsRequestType() cqrs.RequestType { return getTicketQueryType }

type GetTicketResult = dyn.OpResult[domain.Ticket]

var ticketExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "ticketExists"}

type TicketExistsQuery dyn.ExistsQuery

func (TicketExistsQuery) CqrsRequestType() cqrs.RequestType { return ticketExistsQueryType }

type TicketExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "searchTickets"}

type SearchTicketsQuery dyn.SearchQuery

func (SearchTicketsQuery) CqrsRequestType() cqrs.RequestType { return searchTicketsQueryType }

type SearchTicketsResultData = dyn.PagedResultData[domain.Ticket]
type SearchTicketsResult = dyn.OpResult[SearchTicketsResultData]

var updateTicketCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticket", Action: "updateTicket"}

type UpdateTicketCommand struct{ domain.Ticket }

func (UpdateTicketCommand) CqrsRequestType() cqrs.RequestType { return updateTicketCommandType }
func (UpdateTicketCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketSchemaName)
}

type UpdateTicketResult = dyn.OpResult[dyn.MutateResultData]

var setTicketIsArchivedCommandType = cqrs.RequestType{
	Module:    "helpdesk",
	Submodule: "ticket",
	Action:    "setTicketIsArchived",
}

type SetTicketIsArchivedCommand dyn.SetIsArchivedCommand

func (SetTicketIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setTicketIsArchivedCommandType
}

var manageTicketCategoriesCommandType = cqrs.RequestType{
	Module:    "helpdesk",
	Submodule: "ticket",
	Action:    "manageTicketCategories",
}

type ManageTicketCategoriesCommand struct {
	TicketId model.Id                    `json:"ticket_id" param:"ticket_id"`
	Add      datastructure.Set[model.Id] `json:"add"`
	Remove   datastructure.Set[model.Id] `json:"remove"`
}

func (ManageTicketCategoriesCommand) CqrsRequestType() cqrs.RequestType {
	return manageTicketCategoriesCommandType
}

type SetTicketIsArchivedResult = dyn.OpResult[dyn.MutateResultData]
type ManageTicketCategoriesResult = dyn.OpResult[dyn.MutateResultData]
