package app

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
)

// effectiveKey names a setting the way GetEffectiveSettings keys it.
func effectiveKey(name string) string {
	return c.EssentialModuleName + "." + name
}

// NewUserLocaleResolver builds the closure the dynamic-model layer calls to learn which language
// the acting user reads in, so that a search can sort and filter a LangJson column by the text that
// user actually sees.
//
// It lives in Essential because Essential is one of the few packages permitted to name both the
// settings module and core/dynamicmodel. core/dynamicmodel may not name settings at all -- settings
// imports it -- which is why the seam is a function value handed downwards rather than a dependency
// resolved upwards.
func NewUserLocaleResolver(settingsSvc itExt.EffectiveSettingsExtService) dyn.LocaleResolver {
	return func(ctx corectx.Context) *cmodel.LanguageCode {
		// A request with no acting user has no preference to read, and the settings read would only
		// fail trying to find an owner.
		if ctx.GetPermissions().UserId == "" {
			return nil
		}

		result, err := settingsSvc.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{})
		// A settings read that fails must not take the list request down with it: an unlocalized
		// sort is a worse list, a 500 is no list at all.
		if err != nil || result == nil || !result.HasData {
			return defaultLocale()
		}

		// The user's own language first. Falling back to the organization's system_locale keeps a
		// user who never chose one reading in the language their organization works in, which is a
		// better guess than the application default -- but only a fallback, never an override.
		for _, key := range []string{
			effectiveKey(settings.UserSettingLanguage),
			effectiveKey(settings.OrgSettingSystemLocale),
		} {
			if code := supportedLanguageOf(result.Data.Values[key]); code != nil {
				return code
			}
		}
		return defaultLocale()
	}
}

// supportedLanguageOf accepts a stored setting value only if it is a language the application
// actually ships translations for, and returns it in canonical BCP47 form.
//
// The canonicalization is what makes the jsonb lookup correct: LangJson.SanitizeClone rewrites every
// key through the same function on write, so a locale that skipped it would look up a key spelled
// differently from the one stored and silently match nothing.
func supportedLanguageOf(value any) *cmodel.LanguageCode {
	code, ok := value.(string)
	if !ok || !array.Contains(settings.SupportedLanguages, code) {
		return nil
	}
	canonical, err := cmodel.ToBCP47LanguageCode(code)
	if err != nil {
		return nil
	}
	return &canonical
}

// defaultLocale is the locale used when neither the user nor their organization has named one.
//
// It matters far more than it looks: "language" declares no default_value, so a user who has never
// opened the settings pane -- the ordinary case, not an edge case -- has no stored locale anywhere
// in the chain. Returning nil there would leave every LangJson list in the product sorting by the
// raw jsonb document, which is the bug this whole feature exists to fix, for most users.
//
// It is model.DefaultLanguageCode rather than any business-chosen language: BR 53 is explicit that
// vi-VN must not be hardcoded as a default, and this is the display locale of last resort, replaced
// the moment the user or their organization stores a preference.
func defaultLocale() *cmodel.LanguageCode {
	code := cmodel.LanguageCode(cmodel.DefaultLanguageCode)
	return &code
}
