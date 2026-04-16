package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/ticketcategory"
)

type CreateTicketCategoryRequest = it.CreateTicketCategoryCommand
type CreateTicketCategoryResponse = httpserver.RestCreateResponse
type DeleteTicketCategoryRequest = it.DeleteTicketCategoryCommand
type DeleteTicketCategoryResponse = httpserver.RestDeleteResponse2
type GetTicketCategoryRequest = it.GetTicketCategoryQuery
type GetTicketCategoryResponse = dmodel.DynamicFields
type TicketCategoryExistsRequest = it.TicketCategoryExistsQuery
type TicketCategoryExistsResponse = dyn.ExistsResultData
type SearchTicketCategoriesRequest = it.SearchTicketCategoriesQuery
type SearchTicketCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateTicketCategoryRequest = it.UpdateTicketCategoryCommand
type UpdateTicketCategoryResponse = httpserver.RestMutateResponse
type SetTicketCategoryIsArchivedRequest = it.SetTicketCategoryIsArchivedCommand
type SetTicketCategoryIsArchivedResponse = httpserver.RestMutateResponse
