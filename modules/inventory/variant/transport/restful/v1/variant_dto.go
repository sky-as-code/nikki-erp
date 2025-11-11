package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

type VariantDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	ProductId     string  `json:"productId"`
	Sku           string  `json:"sku"`
	Barcode       *string `json:"barcode,omitempty"`
	ProposedPrice *int    `json:"proposedPrice,omitempty"`
	Status        string  `json:"status"`
}

func (this *VariantDto) FromVariant(v it.Variant) {
	model.MustCopy(v.AuditableBase, this)
	model.MustCopy(v.ModelBase, this)
	model.MustCopy(v, this)
}

type CreateVariantRequest = it.CreateVariantCommand
type CreateVariantResponse = httpserver.RestCreateResponse

type UpdateVariantRequest = it.UpdateVariantCommand
type UpdateVariantResponse = httpserver.RestUpdateResponse

type DeleteVariantRequest = it.DeleteVariantCommand
type DeleteVariantResponse = httpserver.RestDeleteResponse

type GetVariantByIdRequest = it.GetVariantByIdQuery
type GetVariantByIdResponse = VariantDto

type SearchVariantsRequest = it.SearchVariantsQuery

type SearchVariantsResponse httpserver.RestSearchResponse[VariantDto]

func (this *SearchVariantsResponse) FromResult(result *it.SearchVariantsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(v it.Variant) VariantDto {
		item := VariantDto{}
		item.FromVariant(v)
		return item
	})
}
