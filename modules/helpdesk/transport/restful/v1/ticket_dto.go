package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticket"
)

type CreateTicketRequest = it.CreateTicketCommand
type CreateTicketResponse = httpserver.RestCreateResponse
type DeleteTicketRequest = it.DeleteTicketCommand
type DeleteTicketResponse = httpserver.RestDeleteResponse2
type GetTicketRequest = it.GetTicketQuery
type GetTicketResponse = dmodel.DynamicFields
type TicketExistsRequest = it.TicketExistsQuery
type TicketExistsResponse = dyn.ExistsResultData
type SearchTicketsRequest = it.SearchTicketsQuery
type SearchTicketsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketRequest = it.UpdateTicketCommand
type UpdateTicketResponse = httpserver.RestMutateResponse
type SetTicketIsArchivedRequest = it.SetTicketIsArchivedCommand
type SetTicketIsArchivedResponse = httpserver.RestMutateResponse
type ManageTicketCategoriesRequest = it.ManageTicketCategoriesCommand
type ManageTicketCategoriesResponse = httpserver.RestMutateResponse
