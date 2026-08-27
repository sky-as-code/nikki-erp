package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Sales' port onto the Settings module's schema registry.
//
// Narrowed to registration alone. Sales declares what an organization may configure about how it
// sells; it never reads or writes another owner's values through this port, and a port that could
// would let a bug here rewrite another module's configuration.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

// EffectiveSettingsExtService is the level-agnostic read Sales uses to resolve a policy.
//
// It answers "what applies to the caller", flattened across tenant, org and user, which is exactly
// the question the pricing and return rules ask. Sales never needs to know which level answered: a
// setting name is unique within a module across all levels, so the levels cannot disagree.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult
