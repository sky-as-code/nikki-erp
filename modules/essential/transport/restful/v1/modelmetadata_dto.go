package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

type CreateModelMetadataRequest struct{ dmodel.DynamicFields }
type CreateModelMetadataResponse = httpserver.RestCreateResponse

type UpdateModelMetadataRequest struct {
	dmodel.DynamicFields
	Id string `json:"id" param:"id"`
}

type UpdateModelMetadataResponse = httpserver.RestMutateResponse
type DeleteModelMetadataRequest = it.DeleteModelMetadataCommand
type DeleteModelMetadataResponse = httpserver.RestDeleteResponse2
type GetModelMetadataRequest = it.GetModelMetadataQuery
type GetModelMetadataResponse = dmodel.DynamicFields
type SearchModelMetadataRequest = it.SearchModelMetadataQuery
type SearchModelMetadataResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type ModelMetadataExistsRequest = it.ModelMetadataExistsQuery
type ModelMetadataExistsResponse = dyn.ExistsResultData
