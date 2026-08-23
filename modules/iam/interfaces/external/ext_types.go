package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	essentialConstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	essentialSettings "github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
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

// UserSettingsExtService seeds a new user's preferences and reads an existing user's.
type UserSettingsExtService interface {
	// InitUserPreferences copies the tenant's settings onto the new user, inside the creating
	// transaction on the same reasoning as InitOrgSettings.
	InitUserPreferences(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)

	// GetUserPreferences reads the acting user's own preferences, which the user-context endpoint
	// reports to the frontend as account settings.
	GetUserPreferences(ctx corectx.Context, query GetSettingsQuery) (*GetSettingsResult, error)
}

// The names below are Essential's, re-exported here so that iam's transport layer does not import
// another module's domain package directly. interfaces/external is the one place iam names a peer
// module, and infra/external is the one place it binds to one.
const (
	// EssentialModuleKey is the module_key Essential registers its user settings under.
	EssentialModuleKey = essentialConstants.EssentialModuleName

	SettingThemeMode = essentialSettings.UserSettingThemeMode
	SettingLanguage  = essentialSettings.UserSettingLanguage

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
