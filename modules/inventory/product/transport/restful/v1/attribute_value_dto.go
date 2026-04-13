package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

type CreateAttributeValueRequest = itAttributeValue.CreateAttributeValueCommand
type CreateAttributeValueResponse = httpserver.RestCreateResponse

type UpdateAttributeValueRequest = itAttributeValue.UpdateAttributeValueCommand
type UpdateAttributeValueResponse = httpserver.RestMutateResponse

type DeleteAttributeValueRequest = itAttributeValue.DeleteAttributeValueCommand
type DeleteAttributeValueResponse = httpserver.RestDeleteResponse2

type GetAttributeValueRequest = itAttributeValue.GetAttributeValueQuery
type GetAttributeValueResponse = dmodel.DynamicFields

type SearchAttributeValuesRequest = itAttributeValue.SearchAttributeValuesQuery
type SearchAttributeValuesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
