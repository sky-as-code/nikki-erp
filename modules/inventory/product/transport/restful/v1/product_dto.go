package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces"
)

type ProductDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	// Optional common product fields (copied if present in domain.Product)
	Name              model.LangJson  `json:"name"`
	Description       *model.LangJson `json:"description,omitempty"`
	Unit              string          `json:"unit_id"`
	Status            string          `json:"status"`
	DefaultsVariantId *string         `json:"defaultsVariantId,omitempty"`
	ThumbnailUrl      *string         `json:"thumbnailUrl,omitempty"`
}

func (this *ProductDto) FromProduct(p it.Product) {
	model.MustCopy(p.AuditableBase, this)
	model.MustCopy(p.ModelBase, this)
	model.MustCopy(p, this)
}

type CreateProductRequest = it.CreateProductCommand
type CreateProductResponse = httpserver.RestCreateResponse

type UpdateProductRequest = it.UpdateProductCommand
type UpdateProductResponse = httpserver.RestUpdateResponse

type DeleteProductRequest = it.DeleteProductCommand
type DeleteProductResponse = httpserver.RestDeleteResponse

type GetProductByIdRequest = it.GetProductByIdQuery
type GetProductByIdResponse = ProductDto

type SearchProductsRequest = it.SearchProductsQuery

type SearchProductsResponse httpserver.RestSearchResponse[ProductDto]

func (this *SearchProductsResponse) FromResult(result *it.SearchProductsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(p it.Product) ProductDto {
		item := ProductDto{}
		item.FromProduct(p)
		return item
	})
}
