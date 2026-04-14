package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

type CreateModuleRequest struct {
	dmodel.DynamicFields
}

type CreateModuleResponse = httpserver.RestCreateResponse

type UpdateModuleRequest struct {
	dmodel.DynamicFields
	Id string `json:"id" param:"id"`
}

type UpdateModuleResponse = httpserver.RestMutateResponse

type DeleteModuleRequest = it.DeleteModuleCommand
type DeleteModuleResponse = httpserver.RestDeleteResponse2

type GetModuleRequest = it.GetModuleQuery
type GetModuleResponse = dmodel.DynamicFields

type SearchModulesRequest = it.SearchModulesQuery
type SearchModulesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type ModuleExistsRequest = it.ModuleExistsQuery
type ModuleExistsResponse = dyn.ExistsResultData
