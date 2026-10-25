package app

import (
	"go.bryk.io/pkg/errors"

	cmodel "github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewEnabledLanguagesResolver builds the closure the dynamic resource engine calls to learn which
// languages a LangJson value may hold for the acting tenant. It lives in Essential for the same
// reason NewUserLocaleResolver does: core/dynamicmodel may not name the language catalogue.
func NewEnabledLanguagesResolver(langSvc itLang.LanguageLookupService) dyn.EnabledLanguagesResolver {
	return func(ctx corectx.Context) ([]cmodel.LanguageCode, error) {
		result, err := langSvc.ListEnabledLanguages(ctx, itLang.ListEnabledLanguagesQuery{})
		if err != nil {
			return nil, errors.Wrap(err, "EnabledLanguagesResolver")
		}
		if result == nil || !result.HasData {
			return nil, nil
		}
		codes := make([]cmodel.LanguageCode, 0, len(result.Data))
		for _, language := range result.Data {
			if code := supportedLanguageOf(safe.GetVal(language.GetIsoCode(), "")); code != nil {
				codes = append(codes, *code)
			}
		}
		return codes, nil
	}
}
