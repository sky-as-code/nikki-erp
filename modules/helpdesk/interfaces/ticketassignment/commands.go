package ticketassignment

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateTicketAssignmentCommand)(nil)
	req = (*DeleteTicketAssignmentCommand)(nil)
	req = (*GetTicketAssignmentQuery)(nil)
	req = (*TicketAssignmentExistsQuery)(nil)
	req = (*SearchTicketAssignmentsQuery)(nil)
	req = (*UpdateTicketAssignmentCommand)(nil)
	util.Unused(req)
}

var createTicketAssignmentCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "createTicketAssignment"}

type CreateTicketAssignmentCommand struct{ models.TicketAssignment }

func (CreateTicketAssignmentCommand) CqrsRequestType() cqrs.RequestType {
	return createTicketAssignmentCommandType
}
func (CreateTicketAssignmentCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TicketAssignmentSchemaName)
}

type CreateTicketAssignmentResult = dyn.OpResult[models.TicketAssignment]

var deleteTicketAssignmentCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "deleteTicketAssignment"}

type DeleteTicketAssignmentCommand dyn.DeleteOneCommand

func (DeleteTicketAssignmentCommand) CqrsRequestType() cqrs.RequestType {
	return deleteTicketAssignmentCommandType
}

type DeleteTicketAssignmentResult = dyn.OpResult[dyn.MutateResultData]

var getTicketAssignmentQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "getTicketAssignment"}

type GetTicketAssignmentQuery dyn.GetOneQuery

func (GetTicketAssignmentQuery) CqrsRequestType() cqrs.RequestType {
	return getTicketAssignmentQueryType
}

type GetTicketAssignmentResult = dyn.OpResult[models.TicketAssignment]

var ticketAssignmentExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "ticketAssignmentExists"}

type TicketAssignmentExistsQuery dyn.ExistsQuery

func (TicketAssignmentExistsQuery) CqrsRequestType() cqrs.RequestType {
	return ticketAssignmentExistsQueryType
}

type TicketAssignmentExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchTicketAssignmentsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "searchTicketAssignments"}

type SearchTicketAssignmentsQuery dyn.SearchQuery

func (SearchTicketAssignmentsQuery) CqrsRequestType() cqrs.RequestType {
	return searchTicketAssignmentsQueryType
}

type SearchTicketAssignmentsResultData = dyn.PagedResultData[models.TicketAssignment]
type SearchTicketAssignmentsResult = dyn.OpResult[SearchTicketAssignmentsResultData]

var updateTicketAssignmentCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "ticketassignment", Action: "updateTicketAssignment"}

type UpdateTicketAssignmentCommand struct{ models.TicketAssignment }

func (UpdateTicketAssignmentCommand) CqrsRequestType() cqrs.RequestType {
	return updateTicketAssignmentCommandType
}
func (UpdateTicketAssignmentCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.TicketAssignmentSchemaName)
}

type UpdateTicketAssignmentResult = dyn.OpResult[dyn.MutateResultData]
