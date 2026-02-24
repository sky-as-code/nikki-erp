package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

type ProductCategoryDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	Name     model.LangJson `json:"name"`
	ParentId *string        `json:"parentId,omitempty"`
}

func (this *ProductCategoryDto) FromProductCategory(pc domain.ProductCategory) {
	model.MustCopy(pc.AuditableBase, this)
	model.MustCopy(pc.ModelBase, this)
	model.MustCopy(pc, this)
}

type CreateProductCategoryRequest = itProductCategory.CreateProductCategoryCommand
type CreateProductCategoryResponse = httpserver.RestCreateResponse

type UpdateProductCategoryRequest = itProductCategory.UpdateProductCategoryCommand
type UpdateProductCategoryResponse = httpserver.RestUpdateResponse

type DeleteProductCategoryRequest = itProductCategory.DeleteProductCategoryCommand
type DeleteProductCategoryResponse = httpserver.RestDeleteResponse

type GetProductCategoryByIdRequest = itProductCategory.GetProductCategoryByIdQuery
type GetProductCategoryByIdResponse = ProductCategoryDto

type SearchProductCategoriesRequest = itProductCategory.SearchProductCategoriesQuery

type SearchProductCategoriesResponse httpserver.RestSearchResponse[ProductCategoryDto]

func (this *SearchProductCategoriesResponse) FromResult(result *itProductCategory.SearchProductCategoriesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(pc domain.ProductCategory) ProductCategoryDto {
		item := ProductCategoryDto{}
		item.FromProductCategory(pc)
		return item
	})
}
