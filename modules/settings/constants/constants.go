package constants

const SettingsModuleName = "settings"

// Setting levels. A level says which kind of owner a setting belongs to, and is copied from the
// registered schema onto every record built from it.
const (
	LevelTenant = "tenant"
	LevelOrg    = "org"
	LevelUser   = "user"
)

// Owner types. They mirror the levels, but describe the row's actual owner rather than the
// schema's declared level: a tenant admin's own record for a user-level setting is
// level=user, owner_type=tenant.
const (
	OwnerTypeTenant = "tenant"
	OwnerTypeOrg    = "org"
	OwnerTypeUser   = "user"
)

// MetadataKeyAllowOverride is the field-metadata key a module sets on a setting field to say
// whether an owner below the tenant may keep its own value. It is declared on the schema field
// rather than stored per record, so it is uniform for a setting name across the whole tenant.
const MetadataKeyAllowOverride = "allow_override"
