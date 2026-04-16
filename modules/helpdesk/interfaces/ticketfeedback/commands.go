package ticketfeedback

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketFeedbackCommand)(nil)
	req = (*DeleteTicketFeedbackCommand)(nil)
	req = (*GetTicketFeedbackQuery)(nil)
	req = (*TicketFeedbackExistsQuery)(nil)
	req = (*SearchTicketFeedbacksQuery)(nil)
	req = (*UpdateTicketFeedbackCommand)(nil)
	util.Unused(req)
}

var createTicketFeedbackCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "createTicketFeedback"}

type CreateTicketFeedbackCommand struct{ domain.TicketFeedback }

func (CreateTicketFeedbackCommand) CqrsRequestType() cqrs.RequestType {
	return createTicketFeedbackCommandType
}
func (CreateTicketFeedbackCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketFeedbackSchemaName)
}

type CreateTicketFeedbackResult = dyn.OpResult[domain.TicketFeedback]

var deleteTicketFeedbackCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "deleteTicketFeedback"}

type DeleteTicketFeedbackCommand dyn.DeleteOneCommand

func (DeleteTicketFeedbackCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTicketFeedbackCommandType
}

type DeleteTicketFeedbackResult = dyn.OpResult[dyn.MutateResultData]

var getTicketFeedbackQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "getTicketFeedback"}

type GetTicketFeedbackQuery dyn.GetOneQuery

func (GetTicketFeedbackQuery) CqrsRequestType() cqrs.RequestType { return getTicketFeedbackQueryType }

type GetTicketFeedbackResult = dyn.OpResult[domain.TicketFeedback]

var ticketFeedbackExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "ticketFeedbackExists"}

type TicketFeedbackExistsQuery dyn.ExistsQuery

func (TicketFeedbackExistsQuery) CqrsRequestType() cqrs.RequestType {
	return ticketFeedbackExistsQueryType
}

type TicketFeedbackExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketFeedbacksQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "searchTicketFeedbacks"}

type SearchTicketFeedbacksQuery dyn.SearchQuery

func (SearchTicketFeedbacksQuery) CqrsRequestType() cqrs.RequestType {
	return searchTicketFeedbacksQueryType
}

type SearchTicketFeedbacksResultData = dyn.PagedResultData[domain.TicketFeedback]
type SearchTicketFeedbacksResult = dyn.OpResult[SearchTicketFeedbacksResultData]

var updateTicketFeedbackCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketfeedback", Action: "updateTicketFeedback"}

type UpdateTicketFeedbackCommand struct{ domain.TicketFeedback }

func (UpdateTicketFeedbackCommand) CqrsRequestType() cqrs.RequestType {
	return updateTicketFeedbackCommandType
}
func (UpdateTicketFeedbackCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketFeedbackSchemaName)
}

type UpdateTicketFeedbackResult = dyn.OpResult[dyn.MutateResultData]
