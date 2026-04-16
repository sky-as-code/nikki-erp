package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketfeedback"
)

type CreateTicketFeedbackRequest = it.CreateTicketFeedbackCommand
type CreateTicketFeedbackResponse = httpserver.RestCreateResponse
type DeleteTicketFeedbackRequest = it.DeleteTicketFeedbackCommand
type DeleteTicketFeedbackResponse = httpserver.RestDeleteResponse2
type GetTicketFeedbackRequest = it.GetTicketFeedbackQuery
type GetTicketFeedbackResponse = dmodel.DynamicFields
type TicketFeedbackExistsRequest = it.TicketFeedbackExistsQuery
type TicketFeedbackExistsResponse = dyn.ExistsResultData
type SearchTicketFeedbacksRequest = it.SearchTicketFeedbacksQuery
type SearchTicketFeedbacksResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketFeedbackRequest = it.UpdateTicketFeedbackCommand
type UpdateTicketFeedbackResponse = httpserver.RestMutateResponse
