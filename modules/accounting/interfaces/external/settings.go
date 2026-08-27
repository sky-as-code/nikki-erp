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
