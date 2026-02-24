package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

type ProductDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	// Optional common product fields (copied if present in domain.Product)
	Name             model.LangJson  `json:"name"`
	Description      *model.LangJson `json:"description,omitempty"`
	UnitId           *model.Id       `json:"unitId,omitempty"`
	Status           string          `json:"status"`
	DefaultVariantId *model.Id       `json:"defaultVariantId,omitempty"`
	ThumbnailUrl     *string         `json:"thumbnailUrl,omitempty"`

	Variants   []GetVariantByProductResponse `json:"variants,omitempty"`
	Attributes []AttributeDto                `json:"attributes,omitempty"`
}

func (this *ProductDto) FromProduct(p domain.Product) {
	model.MustCopy(p.AuditableBase, this)
	model.MustCopy(p.ModelBase, this)
	model.MustCopy(p, this)

	this.Variants = array.Map(p.Variants, func(v domain.Variant) GetVariantByProductResponse {
		variantResp := GetVariantByProductResponse{}
		variantResp.FromVariant(v)
		return variantResp
	})

	this.Attributes = array.Map(p.Attributes, func(attr domain.Attribute) AttributeDto {
		attrResp := AttributeDto{}
		attrResp.FromAttribute(attr)
		return attrResp
	})
}

type GetVariantByProductResponse struct {
	Id            string `json:"id"`
	Sku           string `json:"sku"`
	Barcode       string `json:"barcode,omitempty"`
	ProposedPrice int    `json:"proposedPrice,omitempty"`
	Status        string `json:"status"`
}

func (this *GetVariantByProductResponse) FromVariant(v domain.Variant) {
	model.MustCopy(v.AuditableBase, this)
	model.MustCopy(v.ModelBase, this)
	model.MustCopy(v, this)
}

type CreateProductRequest = itProduct.CreateProductCommand
type CreateProductResponse = httpserver.RestCreateResponse

type UpdateProductRequest = itProduct.UpdateProductCommand
type UpdateProductResponse = httpserver.RestUpdateResponse

type DeleteProductRequest = itProduct.DeleteProductCommand
type DeleteProductResponse = httpserver.RestDeleteResponse

type GetProductByIdRequest = itProduct.GetProductByIdQuery
type GetProductByIdResponse = ProductDto

type SearchProductsRequest = itProduct.SearchProductsQuery

type SearchProductsResponse httpserver.RestSearchResponse[ProductDto]

func (this *SearchProductsResponse) FromResult(result *itProduct.SearchProductsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(p domain.Product) ProductDto {
		item := ProductDto{}
		item.FromProduct(p)
		return item
	})
}
