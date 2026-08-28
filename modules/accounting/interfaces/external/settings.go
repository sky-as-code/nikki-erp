package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Accounting's port onto the Settings module's schema registry.
//
// It is narrowed to registration alone. Accounting declares what an organization may configure
// about its own tax handling; it never reads or writes another owner's values, and an alias of the
// full settings contract would re-export every method added to that contract later.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult

// EffectiveSettingsExtService is Accounting's port onto reading the values of the settings it
// registered.
//
// Separate from SettingsRegistrationExtService above rather than folded into it, because they are
// used at opposite ends of the lifecycle and by different callers: registration happens once at
// boot from OnAppStarted, while a read happens on every request that needs the organization's
// currency. Keeping them apart means the boot path cannot accidentally read and a request path
// cannot accidentally re-register.
type EffectiveSettingsExtService interface {
	GetEffectiveSettings(
		ctx corectx.Context, query GetEffectiveSettingsQuery,
	) (*GetEffectiveSettingsResult, error)
}

type GetEffectiveSettingsQuery = itSettings.GetEffectiveSettingsQuery
type GetEffectiveSettingsResult = itSettings.GetEffectiveSettingsResult
