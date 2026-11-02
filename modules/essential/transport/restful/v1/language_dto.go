package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// ListEnabledLanguagesResponse names only what a language picker shows. It is served without a
// session -- the sign-in page offers it -- so it carries no field a picker does not need.
type ListEnabledLanguagesResponse struct {
	Items []EnabledLanguageDto `json:"items"`
	// DefaultLocaleCode is the system default language: what a user reads when neither they nor
	// their organization chose one. It is the application default when that is enabled, else the
	// first enabled language, and empty only when none is enabled.
	DefaultLocaleCode string `json:"default_locale_code"`
}

type EnabledLanguageDto struct {
	LocaleCode string `json:"locale_code"`
	Name       string `json:"name"`
	Direction  string `json:"direction"`
}

func NewListEnabledLanguagesResponse(languages []models.Language) ListEnabledLanguagesResponse {
	items := make([]EnabledLanguageDto, 0, len(languages))
	for _, language := range languages {
		items = append(items, EnabledLanguageDto{
			LocaleCode: util.ValueOrZeroOf(language.GetLocaleCode()),
			Name:       util.ValueOrZeroOf(language.GetName()),
			Direction:  util.ValueOrZeroOf(language.GetDirection()),
		})
	}
	return ListEnabledLanguagesResponse{Items: items, DefaultLocaleCode: defaultLocaleCodeOf(items)}
}

func defaultLocaleCodeOf(items []EnabledLanguageDto) string {
	for _, item := range items {
		if item.LocaleCode == model.DefaultLanguageCode {
			return item.LocaleCode
		}
	}
	if len(items) > 0 {
		return items[0].LocaleCode
	}
	return ""
}
