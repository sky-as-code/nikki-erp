package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
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
type GetLanguageJsonRequest = it.GetLanguageJsonQuery
type GetLanguageJsonResponse = model.LanguageTranslationItem
type SearchLanguagesRequest = it.SearchLanguagesQuery
type SearchLanguagesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type LanguageExistsRequest = it.LanguageExistsQuery
type LanguageExistsResponse = dyn.ExistsResultData

// ListEnabledLanguagesRequest takes nothing: the list is the same for every caller.
type ListEnabledLanguagesRequest struct{}

// ListEnabledLanguagesResponse names only what a language picker shows. It is served without a
// session -- the sign-in page offers it -- so it carries no field a picker does not need.
type ListEnabledLanguagesResponse struct {
	Items []EnabledLanguageDto `json:"items"`
	// DefaultIsoCode is the system default language: what a user reads when neither they nor
	// their organization chose one. It is the application default when that is enabled, else the
	// first enabled language, and empty only when none is enabled.
	DefaultIsoCode string `json:"default_iso_code"`
}

type EnabledLanguageDto struct {
	IsoCode   string `json:"iso_code"`
	Name      string `json:"name"`
	Direction string `json:"direction"`
}

func NewListEnabledLanguagesResponse(languages []models.Language) ListEnabledLanguagesResponse {
	items := make([]EnabledLanguageDto, 0, len(languages))
	for _, language := range languages {
		items = append(items, EnabledLanguageDto{
			IsoCode:   util.ValueOrZeroOf(language.GetIsoCode()),
			Name:      util.ValueOrZeroOf(language.GetName()),
			Direction: util.ValueOrZeroOf(language.GetDirection()),
		})
	}
	return ListEnabledLanguagesResponse{Items: items, DefaultIsoCode: defaultIsoCodeOf(items)}
}

func defaultIsoCodeOf(items []EnabledLanguageDto) string {
	for _, item := range items {
		if item.IsoCode == model.DefaultLanguageCode {
			return item.IsoCode
		}
	}
	if len(items) > 0 {
		return items[0].IsoCode
	}
	return ""
}
