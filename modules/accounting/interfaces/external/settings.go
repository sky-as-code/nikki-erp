package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Accounting's port onto the Settings module's schema registry,
// narrowed to registration alone so it never reads or writes another owner's values.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult

// EffectiveSettingsExtService reads the values of the settings Accounting registered. It is kept
// separate from registration, which happens once at boot, so the boot path cannot read and a request
// path cannot re-register.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult

// OrgSettingsSeedExtService writes Accounting's own settings for a new organization, from the org
// lifecycle. It writes nothing else: reads go through EffectiveSettingsExtService.
type OrgSettingsSeedExtService interface {
	SeedOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

type SetSettingsCommand = itSettings.SetSettingsCommand
type SetSettingItem = itSettings.SetSettingItem
type SetSettingsResult = itSettings.SetSettingsResult

// OrgSettingsReaderExtService reads the values one organization stores for a module, naming the
// organization explicitly and asserting no permission: the org lifecycle reads the org's language.
type OrgSettingsReaderExtService = coreSettings.OrgSettingsReader

// OrgLanguage is the language an organization's defaults and generated names are written in.
var OrgLanguage = itLang.OrgLanguage
