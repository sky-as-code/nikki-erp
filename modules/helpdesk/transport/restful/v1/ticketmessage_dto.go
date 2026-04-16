package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketmessage"
)

type CreateTicketMessageRequest = it.CreateTicketMessageCommand
type CreateTicketMessageResponse = httpserver.RestCreateResponse
type DeleteTicketMessageRequest = it.DeleteTicketMessageCommand
type DeleteTicketMessageResponse = httpserver.RestDeleteResponse2
type GetTicketMessageRequest = it.GetTicketMessageQuery
type GetTicketMessageResponse = dmodel.DynamicFields
type TicketMessageExistsRequest = it.TicketMessageExistsQuery
type TicketMessageExistsResponse = dyn.ExistsResultData
type SearchTicketMessagesRequest = it.SearchTicketMessagesQuery
type SearchTicketMessagesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketMessageRequest = it.UpdateTicketMessageCommand
type UpdateTicketMessageResponse = httpserver.RestMutateResponse
