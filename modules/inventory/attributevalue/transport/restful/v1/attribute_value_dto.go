package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/interfaces"
)

type AttributeValueDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	AttributeId  string          `json:"attributeId"`
	ValueText    *model.LangJson `json:"valueText,omitempty"`
	ValueNumber  *float64        `json:"valueNumber,omitempty"`
	ValueBool    *bool           `json:"valueBool,omitempty"`
	ValueRef     *string         `json:"valueRef,omitempty"`
	VariantCount *int            `json:"variantCount,omitempty"`
}

func (this *AttributeValueDto) FromAttributeValue(av it.AttributeValue) {
	model.MustCopy(av.AuditableBase, this)
	model.MustCopy(av.ModelBase, this)
	model.MustCopy(av, this)
}

type CreateAttributeValueRequest = it.CreateAttributeValueCommand
type CreateAttributeValueResponse = httpserver.RestCreateResponse

type UpdateAttributeValueRequest = it.UpdateAttributeValueCommand
type UpdateAttributeValueResponse = httpserver.RestUpdateResponse

type DeleteAttributeValueRequest = it.DeleteAttributeValueCommand
type DeleteAttributeValueResponse = httpserver.RestDeleteResponse

type GetAttributeValueByIdRequest = it.GetAttributeValueByIdQuery
type GetAttributeValueByIdResponse = AttributeValueDto

type SearchAttributeValuesRequest = it.SearchAttributeValuesQuery

type SearchAttributeValuesResponse httpserver.RestSearchResponse[AttributeValueDto]

func (this *SearchAttributeValuesResponse) FromResult(result *it.SearchAttributeValuesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(av it.AttributeValue) AttributeValueDto {
		item := AttributeValueDto{}
		item.FromAttributeValue(av)
		return item
	})
}
