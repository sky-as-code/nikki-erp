package ticketmessage

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketMessageCommand)(nil)
	req = (*DeleteTicketMessageCommand)(nil)
	req = (*GetTicketMessageQuery)(nil)
	req = (*TicketMessageExistsQuery)(nil)
	req = (*SearchTicketMessagesQuery)(nil)
	req = (*UpdateTicketMessageCommand)(nil)
	util.Unused(req)
}

var createTicketMessageCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "createTicketMessage"}

type CreateTicketMessageCommand struct{ models.TicketMessage }

func (CreateTicketMessageCommand) CqrsRequestType() cqrs.RequestType {
	return createTicketMessageCommandType
}
func (CreateTicketMessageCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TicketMessageSchemaName)
}

type CreateTicketMessageResult = dyn.OpResult[models.TicketMessage]

var deleteTicketMessageCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "deleteTicketMessage"}

type DeleteTicketMessageCommand dyn.DeleteOneCommand

func (DeleteTicketMessageCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTicketMessageCommandType
}

type DeleteTicketMessageResult = dyn.OpResult[dyn.MutateResultData]

var getTicketMessageQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "getTicketMessage"}

type GetTicketMessageQuery dyn.GetOneQuery

func (GetTicketMessageQuery) CqrsRequestType() cqrs.RequestType { return getTicketMessageQueryType }

type GetTicketMessageResult = dyn.OpResult[models.TicketMessage]

var ticketMessageExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "ticketMessageExists"}

type TicketMessageExistsQuery dyn.ExistsQuery

func (TicketMessageExistsQuery) CqrsRequestType() cqrs.RequestType {
	return ticketMessageExistsQueryType
}

type TicketMessageExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketMessagesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "searchTicketMessages"}

type SearchTicketMessagesQuery dyn.SearchQuery

func (SearchTicketMessagesQuery) CqrsRequestType() cqrs.RequestType {
	return searchTicketMessagesQueryType
}

type SearchTicketMessagesResultData = dyn.PagedResultData[models.TicketMessage]
type SearchTicketMessagesResult = dyn.OpResult[SearchTicketMessagesResultData]

var updateTicketMessageCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketmessage", Action: "updateTicketMessage"}

type UpdateTicketMessageCommand struct{ models.TicketMessage }

func (UpdateTicketMessageCommand) CqrsRequestType() cqrs.RequestType {
	return updateTicketMessageCommandType
}
func (UpdateTicketMessageCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TicketMessageSchemaName)
}

type UpdateTicketMessageResult = dyn.OpResult[dyn.MutateResultData]
