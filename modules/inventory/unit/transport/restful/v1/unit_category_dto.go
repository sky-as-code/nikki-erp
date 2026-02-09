package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

type UnitCategoryDto struct {
	Id           string          `json:"id"`
	CreatedAt    int64           `json:"createdAt"`
	UpdatedAt    *int64          `json:"updatedAt,omitempty"`
	Etag         string          `json:"etag"`
	OrgId        *string         `json:"orgId,omitempty"`
	Name         model.LangJson  `json:"name"`
	Description  *model.LangJson `json:"description,omitempty"`
	Status       *string         `json:"status,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailURL,omitempty"`
}

func (this *UnitCategoryDto) FromUnitCategory(uc domain.UnitCategory) {
	model.MustCopy(uc.AuditableBase, this)
	model.MustCopy(uc.ModelBase, this)
	model.MustCopy(uc, this)
}

type CreateUnitCategoryRequest = itUnitCategory.CreateUnitCategoryCommand
type CreateUnitCategoryResponse = httpserver.RestCreateResponse

type UpdateUnitCategoryRequest = itUnitCategory.UpdateUnitCategoryCommand
type UpdateUnitCategoryResponse = httpserver.RestUpdateResponse

type DeleteUnitCategoryRequest = itUnitCategory.DeleteUnitCategoryCommand
type DeleteUnitCategoryResponse = httpserver.RestDeleteResponse

type GetUnitCategoryByIdRequest = itUnitCategory.GetUnitCategoryByIdQuery
type GetUnitCategoryByIdResponse = UnitCategoryDto

type SearchUnitCategoriesRequest = itUnitCategory.SearchUnitCategoriesQuery

type SearchUnitCategoriesResponse httpserver.RestSearchResponse[UnitCategoryDto]

func (this *SearchUnitCategoriesResponse) FromResult(result *itUnitCategory.SearchUnitCategoriesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(uc domain.UnitCategory) UnitCategoryDto {
		item := UnitCategoryDto{}
		item.FromUnitCategory(uc)
		return item
	})
}
