package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

type CreatePutawayRuleRequest = itWarehouse.CreatePutawayRuleCommand
type CreatePutawayRuleResponse = httpserver.RestCreateResponse

type UpdatePutawayRuleRequest = itWarehouse.UpdatePutawayRuleCommand
type UpdatePutawayRuleResponse = httpserver.RestMutateResponse

type DeletePutawayRuleRequest = itWarehouse.DeletePutawayRuleCommand
type DeletePutawayRuleResponse = httpserver.RestMutateResponse

type SetPutawayRuleArchivedRequest = itWarehouse.SetPutawayRuleArchivedCommand
type SetPutawayRuleArchivedResponse = httpserver.RestMutateResponse

type GetPutawayRuleRequest = itWarehouse.GetPutawayRuleByIdQuery
type GetPutawayRuleResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type PutawayRuleExistsRequest = itWarehouse.PutawayRuleExistsQuery
type PutawayRuleExistsResponse = dyn.ExistsResultData

type SearchPutawayRulesRequest = itWarehouse.SearchPutawayRulesQuery
type SearchPutawayRulesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
