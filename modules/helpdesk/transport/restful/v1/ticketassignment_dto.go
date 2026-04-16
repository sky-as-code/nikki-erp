package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketassignment"
)

type CreateTicketAssignmentRequest = it.CreateTicketAssignmentCommand
type CreateTicketAssignmentResponse = httpserver.RestCreateResponse
type DeleteTicketAssignmentRequest = it.DeleteTicketAssignmentCommand
type DeleteTicketAssignmentResponse = httpserver.RestDeleteResponse2
type GetTicketAssignmentRequest = it.GetTicketAssignmentQuery
type GetTicketAssignmentResponse = dmodel.DynamicFields
type TicketAssignmentExistsRequest = it.TicketAssignmentExistsQuery
type TicketAssignmentExistsResponse = dyn.ExistsResultData
type SearchTicketAssignmentsRequest = it.SearchTicketAssignmentsQuery
type SearchTicketAssignmentsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketAssignmentRequest = it.UpdateTicketAssignmentCommand
type UpdateTicketAssignmentResponse = httpserver.RestMutateResponse
