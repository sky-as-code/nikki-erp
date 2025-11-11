package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/interfaces"
)

type UnitCategoryDto struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt *int64 `json:"updatedAt,omitempty"`
	Etag      string `json:"etag"`

	Name         model.LangJson `json:"name"`
	Description  model.LangJson `json:"description"`
	Status       *string        `json:"status,omitempty"`
	ThumbnailUrl *string        `json:"thumbnailURL,omitempty"`
}

func (this *UnitCategoryDto) FromUnitCategory(uc it.UnitCategory) {
	model.MustCopy(uc.AuditableBase, this)
	model.MustCopy(uc.ModelBase, this)
	model.MustCopy(uc, this)
}

type CreateUnitCategoryRequest = it.CreateUnitCategoryCommand
type CreateUnitCategoryResponse = httpserver.RestCreateResponse

type UpdateUnitCategoryRequest = it.UpdateUnitCategoryCommand
type UpdateUnitCategoryResponse = httpserver.RestUpdateResponse

type DeleteUnitCategoryRequest = it.DeleteUnitCategoryCommand
type DeleteUnitCategoryResponse = httpserver.RestDeleteResponse

type GetUnitCategoryByIdRequest = it.GetUnitCategoryByIdQuery
type GetUnitCategoryByIdResponse = UnitCategoryDto

type SearchUnitCategoriesRequest = it.SearchUnitCategoriesQuery

type SearchUnitCategoriesResponse httpserver.RestSearchResponse[UnitCategoryDto]

func (this *SearchUnitCategoriesResponse) FromResult(result *it.SearchUnitCategoriesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(uc it.UnitCategory) UnitCategoryDto {
		item := UnitCategoryDto{}
		item.FromUnitCategory(uc)
		return item
	})
}
