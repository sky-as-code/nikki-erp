package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Essential's port onto the Settings module's schema registry.
//
// It is narrowed to registration alone. Essential declares what a user may configure about their
// own account; it never reads or writes another owner's values, and an alias of the full settings
// contract would re-export every method added to that contract later.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult

// OrgSettingsSeedExtService writes Essential's own settings for a new organization, from the org
// lifecycle. It writes nothing else.
type OrgSettingsSeedExtService interface {
	SeedOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

type SetSettingsCommand = itSettings.SetSettingsCommand
type SetSettingItem = itSettings.SetSettingItem
type SetSettingsResult = itSettings.SetSettingsResult
