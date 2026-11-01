package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
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

// LanguageExtService resolves the locale code a settings value carries into the formatting rules the
// frontend renders numbers and dates with, and lists the languages a user may pick.
type LanguageExtService interface {
	GetLanguageByLocaleCode(
		ctx corectx.Context, query GetLanguageByLocaleCodeQuery,
	) (*GetLanguageByLocaleCodeResult, error)
	ListEnabledLanguages(
		ctx corectx.Context, query ListEnabledLanguagesQuery,
	) (*ListEnabledLanguagesResult, error)
}

// CurrencyExtService resolves the organization's system currency, which accounting stores as an
// essential_currency id, into what it takes to render an amount in it.
type CurrencyExtService interface {
	GetCurrency(ctx corectx.Context, query GetCurrencyQuery) (*GetCurrencyResult, error)
}

// ModuleCatalogExtService lists the modules the org lifecycle must reach, grouped in dependency
// waves: every module of a wave depends only on modules of earlier waves.
type ModuleCatalogExtService interface {
	ListModuleWaves(ctx corectx.Context) ([][]string, error)
}

// OrgNotificationExtService raises an in-app notification. A refused or undeliverable
// notification is reported as an error for logging; it never changes the outcome of the work it
// reports on.
type OrgNotificationExtService interface {
	NotifyUsers(ctx corectx.Context, cmd NotifyUsersCommand) error
}

type NotifyUsersCommand struct {
	// CommandId makes a redelivered command raise one notification, not two.
	CommandId        string         `json:"command_id"`
	OrgId            model.Id       `json:"org_id"`
	RecipientUserIds []model.Id     `json:"recipient_user_ids"`
	Title            string         `json:"title"`
	Message          string         `json:"message"`
	Severity         string         `json:"severity"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// The names below are Essential's, re-exported here so that iam's transport layer does not import
// another module's domain package directly. interfaces/external is the one place iam names a peer
// module, and infra/external is the one place it binds to one.
const (
	// EssentialModuleKey is the module_key Essential registers its user settings under.
	EssentialModuleKey = essentialConstants.EssentialModuleName

	SettingThemeMode = essentialSettings.UserSettingThemeMode
	SettingDensity   = essentialSettings.UserSettingDensity
	SettingLanguage  = essentialSettings.UserSettingLanguage

	// SettingSystemLanguage is org-level: the language an organization produces its shared
	// documents in.
	SettingSystemLanguage = essentialSettings.OrgSettingSystemLanguage

	SettingTimezone       = essentialSettings.UserSettingTimezone
	SettingSystemTimezone = essentialSettings.OrgSettingSystemTimezone
)

// Accounting's system currency, named by value rather than imported: iam does not depend on
// accounting, it only reads the stored setting, and reads nothing when accounting is not loaded.
// The org value wins over the tenant's. Keep in step with accounting/domain/settings.
const (
	AccountingModuleKey         = "accounting"
	SettingSystemCurrency       = "system_currency"
	SettingTenantSystemCurrency = "tenant_system_currency"

	ThemeModeAuto  = essentialSettings.ThemeModeAuto
	DefaultDensity = essentialSettings.DefaultDensity
)

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
type InitOwnerSettingsCommand = itSettings.InitOwnerSettingsCommand
type InitOwnerSettingsResult = itSettings.InitOwnerSettingsResult
type GetSettingsQuery = itSettings.GetSettingsQuery
type GetSettingsResult = itSettings.GetSettingsResult
type SettingItem = itSettings.SettingItem
type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult

type GetLanguageByLocaleCodeQuery = essentialLanguage.GetLanguageByLocaleCodeQuery
type GetLanguageByLocaleCodeResult = essentialLanguage.GetLanguageByLocaleCodeResult
type ListEnabledLanguagesQuery = essentialLanguage.ListEnabledLanguagesQuery
type ListEnabledLanguagesResult = essentialLanguage.ListEnabledLanguagesResult

type GetCurrencyQuery = essentialCurrency.GetCurrencyQuery
type GetCurrencyResult = essentialCurrency.GetCurrencyResult

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
