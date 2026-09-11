package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type CreateUomRequest = it.CreateUomCommand
type CreateUomResponse = httpserver.RestCreateResponse

type UpdateUomRequest = it.UpdateUomCommand
type UpdateUomResponse = httpserver.RestMutateResponse

type DeleteUomRequest = it.DeleteUomCommand
type DeleteUomResponse = httpserver.RestMutateResponse

type SetUomArchivedRequest = it.SetUomArchivedCommand
type SetUomArchivedResponse = httpserver.RestMutateResponse

type GetUomRequest = it.GetUomByIdQuery
type GetUomResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type UomExistsRequest = it.UomExistsQuery
type UomExistsResponse = dyn.ExistsResultData

type SearchUomsRequest = it.SearchUomsQuery
type SearchUomsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
