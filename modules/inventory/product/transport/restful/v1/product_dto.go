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
	Name              model.LangJson  `json:"name"`
	Description       *model.LangJson `json:"description,omitempty"`
	UnitId            string          `json:"unitId"`
	Status            string          `json:"status"`
	DefaultsVariantId *string         `json:"defaultsVariantId,omitempty"`
	ThumbnailUrl      *string         `json:"thumbnailUrl,omitempty"`
}

func (this *ProductDto) FromProduct(p domain.Product) {
	model.MustCopy(p.AuditableBase, this)
	model.MustCopy(p.ModelBase, this)
	model.MustCopy(p, this)
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
