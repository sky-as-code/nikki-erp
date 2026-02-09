package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttributeGroup "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributegroup"
)

type AttributeGroupDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	// AttributeGroup specific fields
	Name      model.LangJson `json:"name"`
	Index     int            `json:"index"`
	ProductId *string        `json:"productId,omitempty"`
}

func (this *AttributeGroupDto) FromAttributeGroup(ag domain.AttributeGroup) {
	model.MustCopy(ag.AuditableBase, this)
	model.MustCopy(ag.ModelBase, this)
	model.MustCopy(ag, this)
}

type CreateAttributeGroupRequest = itAttributeGroup.CreateAttributeGroupCommand
type CreateAttributeGroupResponse = httpserver.RestCreateResponse

type UpdateAttributeGroupRequest = itAttributeGroup.UpdateAttributeGroupCommand
type UpdateAttributeGroupResponse = httpserver.RestUpdateResponse

type DeleteAttributeGroupRequest = itAttributeGroup.DeleteAttributeGroupCommand
type DeleteAttributeGroupResponse = httpserver.RestDeleteResponse

type GetAttributeGroupByIdRequest = itAttributeGroup.GetAttributeGroupByIdQuery
type GetAttributeGroupByIdResponse = AttributeGroupDto

type SearchAttributeGroupsRequest = itAttributeGroup.SearchAttributeGroupsQuery

type SearchAttributeGroupsResponse httpserver.RestSearchResponse[AttributeGroupDto]

func (this *SearchAttributeGroupsResponse) FromResult(result *itAttributeGroup.SearchAttributeGroupsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(ag domain.AttributeGroup) AttributeGroupDto {
		item := AttributeGroupDto{}
		item.FromAttributeGroup(ag)
		return item
	})
}
