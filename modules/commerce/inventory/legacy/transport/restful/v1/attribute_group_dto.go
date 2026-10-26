package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/interfaces/attributegroup"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateAttributeGroupRequest = itAttributeGroup.CreateAttributeGroupCommand
type CreateAttributeGroupResponse = httpserver.RestCreateResponse

type UpdateAttributeGroupRequest = itAttributeGroup.UpdateAttributeGroupCommand
type UpdateAttributeGroupResponse = httpserver.RestMutateResponse

type DeleteAttributeGroupRequest = itAttributeGroup.DeleteAttributeGroupCommand
type DeleteAttributeGroupResponse = httpserver.RestDeleteResponse2

type GetAttributeGroupRequest = itAttributeGroup.GetAttributeGroupQuery
type GetAttributeGroupResponse = dmodel.DynamicFields

type SearchAttributeGroupsRequest = itAttributeGroup.SearchAttributeGroupsQuery
type SearchAttributeGroupsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
