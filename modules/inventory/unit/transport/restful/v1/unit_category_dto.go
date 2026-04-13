package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

type CreateUnitCategoryRequest = itUnitCategory.CreateUnitCategoryCommand
type CreateUnitCategoryResponse = httpserver.RestCreateResponse

type UpdateUnitCategoryRequest = itUnitCategory.UpdateUnitCategoryCommand
type UpdateUnitCategoryResponse = httpserver.RestMutateResponse

type DeleteUnitCategoryRequest = itUnitCategory.DeleteUnitCategoryCommand
type DeleteUnitCategoryResponse = httpserver.RestDeleteResponse2

type GetUnitCategoryRequest = itUnitCategory.GetUnitCategoryQuery
type GetUnitCategoryResponse = dmodel.DynamicFields

type SearchUnitCategoriesRequest = itUnitCategory.SearchUnitCategoriesQuery
type SearchUnitCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
