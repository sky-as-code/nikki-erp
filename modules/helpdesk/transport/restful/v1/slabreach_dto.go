package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slabreach"
)

type CreateSlaBreachRequest = it.CreateSlaBreachCommand
type CreateSlaBreachResponse = httpserver.RestCreateResponse
type DeleteSlaBreachRequest = it.DeleteSlaBreachCommand
type DeleteSlaBreachResponse = httpserver.RestDeleteResponse2
type GetSlaBreachRequest = it.GetSlaBreachQuery
type GetSlaBreachResponse = dmodel.DynamicFields
type SlaBreachExistsRequest = it.SlaBreachExistsQuery
type SlaBreachExistsResponse = dyn.ExistsResultData
type SearchSlaBreachesRequest = it.SearchSlaBreachesQuery
type SearchSlaBreachesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type UpdateSlaBreachRequest = it.UpdateSlaBreachCommand
type UpdateSlaBreachResponse = httpserver.RestMutateResponse
