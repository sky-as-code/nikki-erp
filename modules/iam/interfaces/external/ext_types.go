package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	essentialConstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	essentialSettings "github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	essentialCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	essentialLanguage "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
	// iamModels "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	// itGrp "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	// itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	// itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	// itUsr "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// The ports below are narrowed local interfaces rather than aliases of the settings contracts.
// An alias would re-export every method added to those contracts later, which is how a consumer
// ends up depending on capabilities it never asked for.
//
// The direction matters: iam depends on settings, never the reverse. Settings may not import iam
// at all — the reverse edge would be a startup-aborting cycle — so organization and user creation
// call *into* settings from here.

// SettingsRegistrationExtService is iam's port onto the Settings module's schema registry.
//
// Narrowed to registration alone: iam declares what a tenant may configure about authentication,
// and never reads or writes another owner's values through this port.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

// OrgSettingsInitExtService seeds a newly created organization with its own settings rows.
type OrgSettingsInitExtService interface {
	// InitOrgSettings copies the tenant's settings onto the new organization. It is called inside
	// the transaction that creates the organization, so that an organization is never visible
	// without the settings it is supposed to have.
	InitOrgSettings(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)
}

// UserSettingsExtService seeds a new user's preferences.
//
// Reading them back is not here: the user-context endpoint wants every level folded together, not
// the user's own layer alone, so it goes through EffectiveSettingsExtService instead.
type UserSettingsExtService interface {
	// InitUserPreferences copies the tenant's settings onto the new user, inside the creating
	// transaction on the same reasoning as InitOrgSettings.
	InitUserPreferences(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)
}

// EffectiveSettingsExtService is iam's port onto the Settings module's level-agnostic read.
//
// The user-context endpoint needs one consolidated answer to "what settings apply to this user",
// folded across tenant, org and user, rather than three reads it would have to fold itself.
// Narrowed to that single read on the same reasoning as SettingsRegistrationExtService: a port that
// could also write would hand that reach to every holder of it.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

// LanguageExtService resolves the iso code a settings value carries into the formatting rules the
// frontend renders numbers and dates with.
type LanguageExtService interface {
	GetLanguageByIsoCode(
		ctx corectx.Context, query GetLanguageByIsoCodeQuery,
	) (*GetLanguageByIsoCodeResult, error)
}

// CurrencyExtService resolves the organization's configured currency into what it takes to render
// an amount in it.
//
// By code rather than by id because essential's `default_currency` setting stores the alphabetic
// code -- unlike accounting's own `org_default_currency`, which stores an id.
type CurrencyExtService interface {
	GetCurrencyByCode(ctx corectx.Context, query GetCurrencyByCodeQuery) (*GetCurrencyByCodeResult, error)
}

// The names below are Essential's, re-exported here so that iam's transport layer does not import
// another module's domain package directly. interfaces/external is the one place iam names a peer
// module, and infra/external is the one place it binds to one.
const (
	// EssentialModuleKey is the module_key Essential registers its user settings under.
	EssentialModuleKey = essentialConstants.EssentialModuleName

	SettingThemeMode = essentialSettings.UserSettingThemeMode
	SettingLanguage  = essentialSettings.UserSettingLanguage

	// SettingSystemLanguage and SettingDefaultCurrency are org-level: the language an organization
	// produces its shared documents in, and the currency it denominates amounts in.
	SettingSystemLanguage  = essentialSettings.OrgSettingSystemLanguage
	SettingDefaultCurrency = essentialSettings.OrgSettingDefaultCurrency

	ThemeModeAuto = essentialSettings.ThemeModeAuto
)

// SupportedLanguages is the locale set the application ships translations for.
func SupportedLanguages() []string {
	return essentialSettings.SupportedLanguages
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
type InitOwnerSettingsCommand = itSettings.InitOwnerSettingsCommand
type InitOwnerSettingsResult = itSettings.InitOwnerSettingsResult
type GetSettingsQuery = itSettings.GetSettingsQuery
type GetSettingsResult = itSettings.GetSettingsResult
type SettingItem = itSettings.SettingItem
type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult

type GetLanguageByIsoCodeQuery = essentialLanguage.GetLanguageByIsoCodeQuery
type GetLanguageByIsoCodeResult = essentialLanguage.GetLanguageByIsoCodeResult

type GetCurrencyByCodeQuery = essentialCurrency.GetCurrencyByCodeQuery
type GetCurrencyByCodeResult = essentialCurrency.GetCurrencyByCodeResult

// const (
// 	UserStatusInvited = iamModels.UserStatusInvited
// 	UserStatusActive  = iamModels.UserStatusActive
// )

// type UserExtService = itUsr.UserDomainService
// type OrganizationExtService = itOrg.OrganizationDomainService
// type OrgUnitExtService = itOrgUnit.OrgUnitDomainService
// type GroupExtService = itGrp.GroupDomainService

// type GetOrgQuery = itOrg.GetOrgQuery
// type GetUserQuery = itUsr.GetUserQuery

// type Organization = iamModels.Organization
