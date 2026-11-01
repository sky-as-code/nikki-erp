package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCore "github.com/sky-as-code/nikki-erp/modules/core/language/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
)

type userRestParams struct {
	dig.In

	SettingsSvc itExt.EffectiveSettingsExtService
	LanguageSvc itExt.LanguageExtService
	CurrencySvc itExt.CurrencyExtService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		SettingsSvc: params.SettingsSvc,
		LanguageSvc: params.LanguageSvc,
		CurrencySvc: params.CurrencySvc,
	}
}

type UserRest struct {
	SettingsSvc itExt.EffectiveSettingsExtService
	LanguageSvc itExt.LanguageExtService
	CurrencySvc itExt.CurrencyExtService
}

func (this UserRest) GetUserContext(echoCtx *echo.Context) (err error) {
	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	userPerm := reqCtx.GetPermissions()
	user := models.NewUserFrom(reqCtx.GetUser())
	echoCtx.JSON(http.StatusOK, GetUserContextResponse{
		Id:           string(user.MustGetId()),
		AvatarFileId: idToStringPtr(user.GetAvatarFileId()),
		DisplayName:  user.MustGetDisplayName(),
		Email:        user.MustGetEmail(),
		Entitlements: userPerm.Entitlements.ToSlice(),

		IsOwner: userPerm.IsOwner,
		UserOrgIds: array.Map(userPerm.UserOrgIds.ToSlice(), func(id model.Id) string {
			return string(id)
		}),
		OrgUnitId:    idToStringPtr(userPerm.OrgUnitId),
		OrgUnitOrgId: idToStringPtr(userPerm.OrgUnitOrgId),

		Orgs: array.Map(user.GetOrgs(), func(org models.Organization) dmodel.DynamicFields {
			return org.GetFieldData()
		}),
		AccountSettings: this.accountSettings(reqCtx),
		SystemSettings: map[string]any{
			"app_name": "Nikki ERP",
		},
	})
	return nil
}

// idToStringPtr carries an optional id across to the wire, keeping absence as null rather than
// collapsing it to an empty string: the frontend's evaluator treats "no org unit" and "an org
// unit whose id is blank" differently.
func idToStringPtr(id *model.Id) *string {
	if id == nil {
		return nil
	}
	return util.ToPtr(string(*id))
}

// accountSettings reports the settings the acting user's session runs on.
//
// Every value is read through the Settings module's consolidated read rather than level by level:
// one call answers "what applies to this user", already folded across tenant, org and user. This
// handler then resolves the two references in that answer -- a language locale code and a currency
// id -- against Essential's catalogues, so the frontend receives records rather than codes it
// would have to look up itself.
//
// It stays fail-soft throughout. The user context is what the whole application boots from, so a
// settings or catalogue read that fails degrades the entry rather than the session: a user who
// cannot load their currency should still be able to work.
func (this UserRest) accountSettings(reqCtx corectx.Context) map[string]any {
	enabled := this.enabledLanguageCodes(reqCtx)
	settings := map[string]any{
		"language":            fallbackLanguage(),
		"timezone":            "Asia/Ho_Chi_Minh",
		"supported_languages": enabled,
		"theme_mode":          itExt.ThemeModeAuto,
		"density":             itExt.DefaultDensity,
	}

	result, err := this.SettingsSvc.GetEffectiveSettings(reqCtx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{itExt.AccountingModuleKey},
	})
	if err != nil || result == nil || !result.HasData {
		return settings
	}
	values := result.Data.Values

	if mode, ok := values[essentialKey(itExt.SettingThemeMode)].(string); ok {
		settings["theme_mode"] = mode
	}
	if density, ok := values[essentialKey(itExt.SettingDensity)].(string); ok {
		settings["density"] = density
	}
	if timezone := firstNonEmptyString(values,
		essentialKey(itExt.SettingTimezone), essentialKey(itExt.SettingSystemTimezone)); timezone != "" {
		settings["timezone"] = timezone
	}
	if language := this.resolveLanguage(reqCtx, values, enabled); language != nil {
		settings["language"] = *language
	}
	if currency := this.resolveCurrency(reqCtx, values); currency != nil {
		settings["currency"] = currency
	}
	return settings
}

// essentialKey names a setting the way GetEffectiveSettings keys it.
func essentialKey(name string) string {
	return itExt.EssentialModuleKey + "." + name
}

// resolveLanguage reads the acting user's display language out of the effective settings and
// resolves it against essential_languages.
//
// The fallback chain -- the user's own language, then the organization's system_language -- is
// deliberately the same one essential/app.NewUserLocaleResolver walks, so that the language a list
// is sorted in is the language the frontend formats it in. Keep the two in step; they are separate
// because that resolver yields a language code for the search layer while this needs the whole
// formatting record.
func (this UserRest) resolveLanguage(
	ctx corectx.Context, values map[string]any, enabled []string,
) *itCore.Language {
	for _, key := range []string{
		essentialKey(itExt.SettingLanguage),
		essentialKey(itExt.SettingSystemLanguage),
	} {
		localeCode, ok := values[key].(string)
		if !ok || !array.Contains(enabled, localeCode) {
			continue
		}
		if language := this.loadLanguage(ctx, localeCode); language != nil {
			return language
		}
	}
	return nil
}

// enabledLanguageCodes lists the languages a user may pick. A failed read degrades to the fallback
// language alone rather than an empty list: the frontend initialises i18n from this list, and an
// empty one would leave it with nothing to render.
func (this UserRest) enabledLanguageCodes(ctx corectx.Context) []string {
	fallback := []string{model.DefaultLanguageCode}
	found, err := this.LanguageSvc.ListEnabledLanguages(ctx, itExt.ListEnabledLanguagesQuery{})
	if err != nil || found == nil || !found.HasData || len(found.Data) == 0 {
		return fallback
	}
	codes := make([]string, 0, len(found.Data))
	for _, language := range found.Data {
		if code := language.GetLocaleCode(); code != nil {
			codes = append(codes, *code)
		}
	}
	return codes
}

func (this UserRest) loadLanguage(ctx corectx.Context, localeCode string) *itCore.Language {
	found, err := this.LanguageSvc.GetLanguageByLocaleCode(ctx, itExt.GetLanguageByLocaleCodeQuery{
		LocaleCode: localeCode,
	})
	if err != nil || found == nil || !found.HasData {
		return nil
	}
	record := found.Data
	return &itCore.Language{
		Id:                 record.GetId(),
		Name:               record.GetName(),
		LocaleCode:         util.ToPtr(model.LanguageCode(util.ValueOrZeroOf(record.GetLocaleCode()))),
		Direction:          record.GetDirection(),
		DecimalSeparator:   record.GetDecimalSeparator(),
		ThousandsSeparator: record.GetThousandsSeparator(),
		Grouping:           record.GetGrouping(),
		DateFormat:         record.GetDateFormat(),
		TimeFormat:         record.GetTimeFormat(),
		ShortTimeFormat:    record.GetShortTimeFormat(),
		FirstDayOfWeek:     record.GetFirstDayOfWeek(),
	}
}

// resolveCurrency resolves accounting's system currency: the organization's own, else the tenant's.
//
// An absent or unresolvable currency omits the key rather than guessing one. There is no safe
// default currency: a guessed one silently reinterprets every amount on screen.
func (this UserRest) resolveCurrency(ctx corectx.Context, values map[string]any) map[string]any {
	currencyId := firstNonEmptyString(values,
		accountingKey(itExt.SettingSystemCurrency), accountingKey(itExt.SettingTenantSystemCurrency))
	if currencyId == "" {
		return nil
	}
	found, err := this.CurrencySvc.GetCurrency(ctx, itExt.GetCurrencyQuery{Id: model.Id(currencyId)})
	if err != nil || found == nil || !found.HasData {
		return nil
	}
	return map[string]any{
		"code": found.Data.Code,
		// Only the currencies in use are seeded with a symbol, so the frontend renders the code
		// whenever symbol is empty.
		"symbol":          found.Data.Symbol,
		"symbol_position": found.Data.SymbolPosition,
		"decimal_places":  found.Data.DecimalPlaces,
	}
}

func accountingKey(name string) string {
	return itExt.AccountingModuleKey + "." + name
}

// firstNonEmptyString returns the first of keys holding a non-empty string in values.
func firstNonEmptyString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, _ := values[key].(string); value != "" {
			return value
		}
	}
	return ""
}

// fallbackLanguage is the display language of last resort, used only when the settings read fails
// outright or names a language essential_languages cannot resolve.
//
// It is a literal because the response must always carry a language: the frontend formats every
// number and date from it, and a session that booted without one would render nothing at all. It is
// en-US rather than any business-chosen language -- BR §53 is explicit that vi-VN must not be
// hardcoded as a default -- and it is replaced the moment a real record resolves.
func fallbackLanguage() itCore.Language {
	return itCore.Language{
		Name:               util.ToPtr("English (US)"),
		LocaleCode:         util.ToPtr(model.LanguageCode(model.DefaultLanguageCode)),
		Direction:          util.ToPtr("ltr"),
		DecimalSeparator:   util.ToPtr("."),
		ThousandsSeparator: util.ToPtr(","),
		Grouping:           util.ToPtr("[3,0]"),
		DateFormat:         util.ToPtr("%m/%d/%Y"),
		TimeFormat:         util.ToPtr("%H:%M:%S"),
		ShortTimeFormat:    util.ToPtr("%H:%M"),
		FirstDayOfWeek:     util.ToPtr("sunday"),
	}
}
