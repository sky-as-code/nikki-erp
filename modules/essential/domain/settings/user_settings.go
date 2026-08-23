// Package settings holds Essential's settings schemas.
//
// These are separate from domain/models because they are a different kind of thing: a model schema
// maps to a database table this module owns, while a settings schema is metadata only — it owns no
// table, and its values are stored as settings_records rows by the settings module.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Setting names Essential declares at the user level. They are referenced by the settings module
// and by whatever reads a user's theme or locale, so they are constants rather than literals.
const (
	UserSettingThemeMode = "theme_mode"
	UserSettingLanguage  = "language"
	UserSettingTimezone  = "timezone"
)

// Setting names Essential declares at the organization level. These describe how an organization
// presents data to everyone working in it — which locale, which clock, which currency — so they
// belong to the organization rather than to the person reading the screen.
const (
	OrgSettingSystemLocale    = "system_locale"
	OrgSettingSystemTimezone  = "system_timezone"
	OrgSettingDefaultCurrency = "default_currency"
)

// Theme modes. "auto" follows the operating system rather than pinning a choice, and it is the
// default because a user who has expressed no preference has not asked for either fixed theme.
const (
	ThemeModeLight = "light"
	ThemeModeDark  = "dark"
	ThemeModeAuto  = "auto"
)

// SupportedLanguages is the set of locales the application ships translations for.
//
// A value outside this set is rejected rather than stored: an unsupported locale renders every key
// as its raw namespace:key, so accepting one would break the interface for the user who chose it.
// There is no fallback language, which is what makes this list load-bearing rather than advisory.
//
// The same list is spelled out as the enum values of `language` in user_settings.json and of
// `system_locale` in org_settings.json. The duplication is deliberate — the JSON is the schema and
// this is what Go code validates against — and the smoke tests assert the two stay equal.
var SupportedLanguages = []string{"en-US", "vi-VN"}

// UserSettingsSchemaName is the name Essential registers its user-level settings under. It is not
// a table: the schema describes values that settings_records stores, and only the settings module
// owns tables.
const UserSettingsSchemaName = "essential_user_settings"

// OrgSettingsSchemaName is the name Essential registers its organization-level settings under.
// Like the user-level schema it is metadata only and owns no table.
const OrgSettingsSchemaName = "essential_org_settings"

//go:embed user_settings.json
var userSettingsSchemaJson string

//go:embed org_settings.json
var orgSettingsSchemaJson string

// UserSettingsSchemaBuilder declares the settings every user may set for themselves.
//
// The document declares no table_name and no should_build_db: a settings schema is metadata only,
// validated and rendered from, never persisted as a table of its own. It extends no base model
// either — the core.basemodel mixins inject tenant_id and audit columns, which a schema with no
// table must not carry.
//
// Every field carries allow_override: true. A tenant may still enforce any of them — that is the
// tenant administrator's choice, made per tenant in the settings UI — but none is something the
// product locks by default. Theme is a personal accessibility choice, language decides whether a
// user can read the interface at all, and a timezone is where the person actually is.
//
// Timezone is free text rather than an enum: the IANA zone database ships hundreds of names and
// gains more over time, so pinning a list here would reject a valid zone the moment the database
// moved ahead of this declaration.
func UserSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(userSettingsSchemaJson)
}

// OrgSettingsSchemaBuilder declares what an organization configures for everyone working in it.
//
// These are organization-level rather than user-level because they describe the organization's own
// conventions — the locale its documents are written in, the clock its business day runs on, the
// currency its figures are quoted in. A user may sit in a different timezone from their
// organization, which is why UserSettingTimezone exists alongside OrgSettingSystemTimezone: the
// organization's is what a shared report is stamped with, the user's is how that report is shown
// to the person reading it.
//
// system_locale is an enum where system_timezone is not, on the same reasoning as the user-level
// schema: the supported locale set is exactly the set of translations the application ships.
//
// default_currency is free text rather than a reference to essential_currency: a settings schema
// declares no edges, and the settings module stores values as scalars. Validating the code against
// the currency table belongs to whatever spends money, not to the settings pane.
//
// All three carry allow_override: true, so a tenant may standardise them across its organizations
// without the product forcing that choice.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
