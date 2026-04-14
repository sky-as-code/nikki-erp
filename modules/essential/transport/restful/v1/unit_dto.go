package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unit"
)

type CreateUnitRequest = it.CreateUnitCommand
type CreateUnitResponse = httpserver.RestCreateResponse

type UpdateUnitRequest = it.UpdateUnitCommand
type UpdateUnitResponse = httpserver.RestMutateResponse

type DeleteUnitRequest = it.DeleteUnitCommand
type DeleteUnitResponse = httpserver.RestDeleteResponse2

type GetUnitRequest = it.GetUnitQuery
type GetUnitResponse = dmodel.DynamicFields

type SearchUnitsRequest = it.SearchUnitsQuery
type SearchUnitsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
