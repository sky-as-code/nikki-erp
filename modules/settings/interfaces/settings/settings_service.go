package settings

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type RegisterSchemaResult = dyn.OpResult[RegisterSchemaResultData]
type GetSettingsResult = dyn.OpResult[GetSettingsResultData]
type SetSettingsResult = dyn.OpResult[SetSettingsResultData]
type InitOwnerSettingsResult = dyn.OpResult[InitOwnerSettingsResultData]

// SchemaRegistrationService is how a feature module declares what it can be configured with.
// Registration happens during start-up and is idempotent, so a module calls it unconditionally.
type SchemaRegistrationService interface {
	RegisterSchema(ctx corectx.Context, cmd RegisterSchemaCommand) (*RegisterSchemaResult, error)
}

// SettingsDomainService is the full capability, implemented inside Settings. It performs no
// permission checks: authorization belongs to the application services above it, which know which
// level the caller is acting at.
type SettingsDomainService interface {
	SchemaRegistrationService

	// GetSettings returns every item of a module at the given level for one owner, filling in
	// schema defaults for names that have no row yet.
	GetSettings(ctx corectx.Context, level string, ownerType string, query GetSettingsQuery) (*GetSettingsResult, error)

	// SetSettings writes the changed items of one owner at one level in a single transaction,
	// fanning an enforced tenant value out onto its children.
	SetSettings(ctx corectx.Context, level string, ownerType string, cmd SetSettingsCommand) (*SetSettingsResult, error)

	// InitOwnerSettings copies the tenant's own rows onto a newly created organization or user.
	InitOwnerSettings(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)
}

// TenantSettingsAppService is the tenant-level contract other modules consume.
//
// The three per-level services are deliberately separate rather than one service taking a level
// argument: a consumer holding the org contract must not be able to read or write tenant or user
// rows through it, which a level parameter would allow by passing a different string.
type TenantSettingsAppService interface {
	SchemaRegistrationService

	GetTenantSettings(ctx corectx.Context, query GetSettingsQuery) (*GetSettingsResult, error)
	SetTenantSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

// OrgSettingsAppService is the organization-level contract other modules consume.
type OrgSettingsAppService interface {
	GetOrgSettings(ctx corectx.Context, query GetSettingsQuery) (*GetSettingsResult, error)
	SetOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)

	// InitOrgSettings seeds a newly created organization from its tenant's rows. It is called by
	// iam inside the creating transaction, through a port iam owns.
	InitOrgSettings(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)
}

// UserPreferencesAppService is the user-level contract other modules consume.
type UserPreferencesAppService interface {
	GetUserPreferences(ctx corectx.Context, query GetSettingsQuery) (*GetSettingsResult, error)
	SetUserPreferences(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)

	// InitUserPreferences seeds a newly created user from its tenant's rows, called by iam inside
	// the creating transaction.
	InitUserPreferences(ctx corectx.Context, cmd InitOwnerSettingsCommand) (*InitOwnerSettingsResult, error)
}
