package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

type CreateUomCatRequest = it.CreateUomCatCommand
type CreateUomCatResponse = httpserver.RestCreateResponse

type UpdateUomCatRequest = it.UpdateUomCatCommand
type UpdateUomCatResponse = httpserver.RestMutateResponse

type DeleteUomCatRequest = it.DeleteUomCatCommand
type DeleteUomCatResponse = httpserver.RestMutateResponse

type SetUomCatArchivedRequest = it.SetUomCatArchivedCommand
type SetUomCatArchivedResponse = httpserver.RestMutateResponse

type GetUomCatRequest = it.GetUomCatByIdQuery
type GetUomCatResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type UomCatExistsRequest = it.UomCatExistsQuery
type UomCatExistsResponse = dyn.ExistsResultData

type SearchUomCatsRequest = it.SearchUomCatsQuery
type SearchUomCatsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
