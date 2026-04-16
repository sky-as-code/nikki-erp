package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

type CreateLanguageRequest struct{ dmodel.DynamicFields }
type CreateLanguageResponse = httpserver.RestCreateResponse

type UpdateLanguageRequest struct {
	dmodel.DynamicFields
	Id string `json:"id" param:"id"`
}

type UpdateLanguageResponse = httpserver.RestMutateResponse
type DeleteLanguageRequest = it.DeleteLanguageCommand
type DeleteLanguageResponse = httpserver.RestDeleteResponse2
type GetLanguageRequest = it.GetLanguageQuery
type GetLanguageResponse = dmodel.DynamicFields
type SearchLanguagesRequest = it.SearchLanguagesQuery
type SearchLanguagesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type LanguageExistsRequest = it.LanguageExistsQuery
type LanguageExistsResponse = dyn.ExistsResultData
