package v1

import (
	"encoding/json"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
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
	EnumValue     []interface{}   `json:"enumValue,omitempty"`
	EnumValueSort *bool           `json:"enumValueSort,omitempty"`
	GroupId       *string         `json:"groupId,omitempty"`
	ValuesCount   *int            `json:"valuesCount,omitempty"`
	VariantsCount *int            `json:"variantsCount,omitempty"`

	AttributeValues []AttributeValueDto `json:"attributeValues,omitempty"`
	Variants        []VariantDto        `json:"variants,omitempty"`
}

func (this *AttributeDto) FromAttribute(a domain.Attribute) {
	model.MustCopy(a.AuditableBase, this)
	model.MustCopy(a.ModelBase, this)
	model.MustCopy(a, this)

	if a.EnumValue != nil {
		if *a.DataType == "string" {
			for _, v := range *a.EnumValue {
				var enumValue model.LangJson
				err := json.Unmarshal(v, &enumValue)
				if err == nil {
					this.EnumValue = append(this.EnumValue, enumValue)
				}
			}
		}
		if *a.DataType == "number" {
			for _, v := range *a.EnumValue {
				var enumValue float64
				err := json.Unmarshal(v, &enumValue)
				if err == nil {
					this.EnumValue = append(this.EnumValue, enumValue)
				}
			}
		}
	}
}

type CreateAttributeRequest = itAttribute.CreateAttributeCommand
type CreateAttributeResponse = httpserver.RestCreateResponse

type UpdateAttributeRequest = itAttribute.UpdateAttributeCommand
type UpdateAttributeResponse = httpserver.RestUpdateResponse

type DeleteAttributeRequest = itAttribute.DeleteAttributeCommand
type DeleteAttributeResponse = httpserver.RestDeleteResponse

type GetAttributeByIdRequest = itAttribute.GetAttributeByIdQuery
type GetAttributeByIdResponse = AttributeDto

type SearchAttributesRequest = itAttribute.SearchAttributesQuery

type SearchAttributesResponse httpserver.RestSearchResponse[AttributeDto]

func (this *SearchAttributesResponse) FromResults(result *itAttribute.SearchAttributesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(a domain.Attribute) AttributeDto {
		item := AttributeDto{}
		item.FromAttribute(a)

		if len(a.AttributeValues) > 0 {
			count := len(a.AttributeValues)
			item.ValuesCount = &count
		}

		if len(a.Variants) > 0 {
			count := len(a.Variants)
			item.VariantsCount = &count
		}

		return item
	})
}

func (this *AttributeDto) FromResult(a domain.Attribute) {
	this.FromAttribute(a)

	if len(a.AttributeValues) > 0 {
		this.AttributeValues = array.Map(a.AttributeValues, func(av domain.AttributeValue) AttributeValueDto {
			avDto := AttributeValueDto{}
			avDto.FromAttributeValue(av)
			return avDto
		})
	}

	if len(a.Variants) > 0 {
		this.Variants = array.Map(a.Variants, func(v domain.Variant) VariantDto {
			vDto := VariantDto{}
			vDto.FromVariant(v)
			return vDto
		})
	}
}
