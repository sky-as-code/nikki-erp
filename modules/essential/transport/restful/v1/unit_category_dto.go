package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unitcategory"
)

type CreateUnitCategoryRequest = it.CreateUnitCategoryCommand
type CreateUnitCategoryResponse = httpserver.RestCreateResponse

type UpdateUnitCategoryRequest = it.UpdateUnitCategoryCommand
type UpdateUnitCategoryResponse = httpserver.RestMutateResponse

type DeleteUnitCategoryRequest = it.DeleteUnitCategoryCommand
type DeleteUnitCategoryResponse = httpserver.RestDeleteResponse2

type GetUnitCategoryRequest = it.GetUnitCategoryQuery
type GetUnitCategoryResponse = dmodel.DynamicFields

type SearchUnitCategoriesRequest = it.SearchUnitCategoriesQuery
type SearchUnitCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
