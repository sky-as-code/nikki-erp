package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

type CreateFieldMetadataRequest struct{ dmodel.DynamicFields }
type CreateFieldMetadataResponse = httpserver.RestCreateResponse

type UpdateFieldMetadataRequest struct {
	dmodel.DynamicFields
	Id string `json:"id" param:"id"`
}

type UpdateFieldMetadataResponse = httpserver.RestMutateResponse
type DeleteFieldMetadataRequest = it.DeleteFieldMetadataCommand
type DeleteFieldMetadataResponse = httpserver.RestDeleteResponse2
type GetFieldMetadataRequest = it.GetFieldMetadataQuery
type GetFieldMetadataResponse = dmodel.DynamicFields
type SearchFieldMetadataRequest = it.SearchFieldMetadataQuery
type SearchFieldMetadataResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type FieldMetadataExistsRequest = it.FieldMetadataExistsQuery
type FieldMetadataExistsResponse = dyn.ExistsResultData
