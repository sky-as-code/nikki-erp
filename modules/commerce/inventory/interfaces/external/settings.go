package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// EffectiveSettingsExtService reads the Commerce org settings Inventory acts on, such as the
// operation type a stock correction is recorded under.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult

// OrgSettingsSeedExtService writes, from the org lifecycle, the Commerce settings that point at the
// operation types Inventory just created for a new organization.
type OrgSettingsSeedExtService interface {
	SeedOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

type SetSettingsCommand = itSettings.SetSettingsCommand
type SetSettingItem = itSettings.SetSettingItem
type SetSettingsResult = itSettings.SetSettingsResult

// SettingsRegistrationExtService is Inventory's port onto the Settings module's schema registry,
// narrowed to registration alone so a bug here cannot rewrite another module's configuration.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
