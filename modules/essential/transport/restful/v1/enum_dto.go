package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

type CreateEnumRequest struct {
	dmodel.DynamicFields
}
type CreateEnumResponse = httpserver.RestCreateResponse

type UpdateEnumRequest struct {
	dmodel.DynamicFields
	EnumId string `param:"id"`
}
type UpdateEnumResponse = httpserver.RestMutateResponse

type DeleteEnumRequest = it.DeleteEnumCommand
type DeleteEnumResponse = httpserver.RestDeleteResponse2

type GetEnumRequest = it.GetEnumQuery
type GetEnumResponse = dmodel.DynamicFields

type EnumExistsRequest = it.EnumExistsQuery
type EnumExistsResponse = dyn.ExistsResultData

type SearchEnumsRequest = it.SearchEnumsQuery
type SearchEnumsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
