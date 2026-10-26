package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Sales' port onto the Settings module's schema registry,
// narrowed to registration alone so a bug here cannot rewrite another module's configuration.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

// EffectiveSettingsExtService answers "what applies to the caller", flattened across tenant, org
// and user. Sales never needs to know which level answered: a setting name is unique within a
// module across all levels, so the levels cannot disagree.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult
