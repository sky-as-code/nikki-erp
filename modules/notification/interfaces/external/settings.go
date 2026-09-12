package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SettingsRegistrationExtService is Notification's port onto the Settings module's schema registry.
//
// It is narrowed to registration alone. Notification declares what an organization may configure
// about its own notifications; it never reads or writes another owner's values, and an alias of the
// full settings contract would re-export every method added to that contract later.
type SettingsRegistrationExtService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

// OrgSettingsReadExtService is Notification's port onto reading an organization's own settings.
//
// It is separate from registration, and narrowed to the organization level, for the reason the
// settings module splits its own contracts: a holder of this one must not be able to reach a
// tenant's or a user's rows, which a single service taking a level argument would allow.
//
// The organization is the one the request is acting as, and is never named by the caller -- that
// is what stops a caller reading the configuration of an organization it has no business in.
type OrgSettingsReadExtService interface {
	GetOrgSettings(ctx corectx.Context, query GetSettingsQuery) (*GetSettingsResult, error)
}

type RegisterSchemaCommand = itSettings.RegisterSchemaCommand
type RegisterSchemaResult = itSettings.RegisterSchemaResult
type GetSettingsQuery = itSettings.GetSettingsQuery
type GetSettingsResult = itSettings.GetSettingsResult
type SettingItem = itSettings.SettingItem
