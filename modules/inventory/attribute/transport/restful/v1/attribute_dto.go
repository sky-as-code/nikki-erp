package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/interfaces"
)

type AttributeDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	ProductId     string          `json:"productId"`
	CodeName      string          `json:"codeName"`
	DisplayName   *model.LangJson `json:"displayName,omitempty"`
	SortIndex     *int            `json:"sortIndex,omitempty"`
	DataType      string          `json:"dataType"`
	IsRequired    *bool           `json:"isRequired,omitempty"`
	IsEnum        *bool           `json:"isEnum,omitempty"`
	EnumValue     *model.LangJson `json:"enumValue,omitempty"`
	EnumValueSort *bool           `json:"enumValueSort,omitempty"`
	GroupId       *string         `json:"groupId,omitempty"`
}

func (this *AttributeDto) FromAttribute(a it.Attribute) {
	model.MustCopy(a.AuditableBase, this)
	model.MustCopy(a.ModelBase, this)
	model.MustCopy(a, this)
}

type CreateAttributeRequest = it.CreateAttributeCommand
type CreateAttributeResponse = httpserver.RestCreateResponse

type UpdateAttributeRequest = it.UpdateAttributeCommand
type UpdateAttributeResponse = httpserver.RestUpdateResponse

type DeleteAttributeRequest = it.DeleteAttributeCommand
type DeleteAttributeResponse = httpserver.RestDeleteResponse

type GetAttributeByIdRequest = it.GetAttributeByIdQuery
type GetAttributeByIdResponse = AttributeDto

type SearchAttributesRequest = it.SearchAttributesQuery

type SearchAttributesResponse httpserver.RestSearchResponse[AttributeDto]

func (this *SearchAttributesResponse) FromResult(result *it.SearchAttributesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(a it.Attribute) AttributeDto {
		item := AttributeDto{}
		item.FromAttribute(a)
		return item
	})
}
