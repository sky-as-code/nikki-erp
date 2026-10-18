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
