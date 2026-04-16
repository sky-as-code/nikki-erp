package ticketactivity

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketActivityCommand)(nil)
	req = (*DeleteTicketActivityCommand)(nil)
	req = (*GetTicketActivityQuery)(nil)
	req = (*TicketActivityExistsQuery)(nil)
	req = (*SearchTicketActivitiesQuery)(nil)
	req = (*UpdateTicketActivityCommand)(nil)
	util.Unused(req)
}

var createTicketActivityCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "createTicketActivity"}

type CreateTicketActivityCommand struct{ domain.TicketActivity }

func (CreateTicketActivityCommand) CqrsRequestType() cqrs.RequestType {
	return createTicketActivityCommandType
}
func (CreateTicketActivityCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketActivitySchemaName)
}

type CreateTicketActivityResult = dyn.OpResult[domain.TicketActivity]

var deleteTicketActivityCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "deleteTicketActivity"}

type DeleteTicketActivityCommand dyn.DeleteOneCommand

func (DeleteTicketActivityCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTicketActivityCommandType
}

type DeleteTicketActivityResult = dyn.OpResult[dyn.MutateResultData]

var getTicketActivityQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "getTicketActivity"}

type GetTicketActivityQuery dyn.GetOneQuery

func (GetTicketActivityQuery) CqrsRequestType() cqrs.RequestType { return getTicketActivityQueryType }

type GetTicketActivityResult = dyn.OpResult[domain.TicketActivity]

var ticketActivityExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "ticketActivityExists"}

type TicketActivityExistsQuery dyn.ExistsQuery

func (TicketActivityExistsQuery) CqrsRequestType() cqrs.RequestType {
	return ticketActivityExistsQueryType
}

type TicketActivityExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketActivitiesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "searchTicketActivities"}

type SearchTicketActivitiesQuery dyn.SearchQuery

func (SearchTicketActivitiesQuery) CqrsRequestType() cqrs.RequestType {
	return searchTicketActivitiesQueryType
}

type SearchTicketActivitiesResultData = dyn.PagedResultData[domain.TicketActivity]
type SearchTicketActivitiesResult = dyn.OpResult[SearchTicketActivitiesResultData]

var updateTicketActivityCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketactivity", Action: "updateTicketActivity"}

type UpdateTicketActivityCommand struct{ domain.TicketActivity }

func (UpdateTicketActivityCommand) CqrsRequestType() cqrs.RequestType {
	return updateTicketActivityCommandType
}
func (UpdateTicketActivityCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.TicketActivitySchemaName)
}

type UpdateTicketActivityResult = dyn.OpResult[dyn.MutateResultData]
