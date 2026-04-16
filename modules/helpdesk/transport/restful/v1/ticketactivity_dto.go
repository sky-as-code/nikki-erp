package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketactivity"
)

type CreateTicketActivityRequest = it.CreateTicketActivityCommand
type CreateTicketActivityResponse = httpserver.RestCreateResponse
type DeleteTicketActivityRequest = it.DeleteTicketActivityCommand
type DeleteTicketActivityResponse = httpserver.RestDeleteResponse2
type GetTicketActivityRequest = it.GetTicketActivityQuery
type GetTicketActivityResponse = dmodel.DynamicFields
type TicketActivityExistsRequest = it.TicketActivityExistsQuery
type TicketActivityExistsResponse = dyn.ExistsResultData
type SearchTicketActivitiesRequest = it.SearchTicketActivitiesQuery
type SearchTicketActivitiesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketActivityRequest = it.UpdateTicketActivityCommand
type UpdateTicketActivityResponse = httpserver.RestMutateResponse
