package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
)

type CreateAttributeRequest = itAttribute.CreateAttributeCommand
type CreateAttributeResponse = httpserver.RestCreateResponse

type UpdateAttributeRequest = itAttribute.UpdateAttributeCommand
type UpdateAttributeResponse = httpserver.RestMutateResponse

type DeleteAttributeRequest = itAttribute.DeleteAttributeCommand
type DeleteAttributeResponse = httpserver.RestDeleteResponse2

type GetAttributeRequest = itAttribute.GetAttributeQuery
type GetAttributeResponse = dmodel.DynamicFields

type SearchAttributesRequest = itAttribute.SearchAttributesQuery
type SearchAttributesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
