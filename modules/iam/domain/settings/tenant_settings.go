// Package settings holds iam's settings schemas.
//
// These are separate from domain/models because they are a different kind of thing: a model schema
// maps to a database table this module owns, while a settings schema is metadata only — it owns no
// table, and its values are stored as settings_records rows by the settings module.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Setting names iam declares at the tenant level. Both govern how everyone in the tenant signs in
// and stays signed in, which is a tenant-wide security policy rather than a personal preference.
const (
	TenantSettingSessionTimeoutMinutes = "session_timeout_minutes"
	TenantSettingRequireMfa            = "require_mfa"
)

// Bounds for the session timeout, in minutes, as declared in tenant_settings.json.
//
// The floor is one minute rather than zero because zero would mean "expire immediately", locking
// every account in the tenant out with no way back in through the interface that set it. The
// ceiling is a week: a session that outlives that is indistinguishable from no timeout at all, and
// a tenant wanting that should say so by other means rather than by typing a large number here.
//
// KNOWN LIMITATION: zero is not refused. ModelField.Validate treats a numeric zero as an ABSENT
// value platform-wide — isNilOrEmpty returns true for it — so the range check is skipped before it
// is reached and a submitted 0 is stored as "no value chosen" rather than rejected. Changing that
// helper would alter validation for every numeric field in the product, so it is not done here.
//
// The practical effect is contained: 0 reads back as unset rather than as "expire immediately", so
// no account is locked out. A caller who means "never expire" still cannot say so, which is the
// right answer — the ceiling is how a tenant expresses a very long session.
const (
	SessionTimeoutMinMinutes = int32(1)
	SessionTimeoutMaxMinutes = int32(60 * 24 * 7)
)

// TenantSettingsSchemaName is the name iam registers its tenant-level settings under. It is not a
// table: the schema describes values that settings_records stores, and only the settings module
// owns tables.
const TenantSettingsSchemaName = "iam_tenant_settings"

//go:embed tenant_settings.json
var tenantSettingsSchemaJson string

// TenantSettingsSchemaBuilder declares the security policy a tenant sets for everyone in it.
//
// The document declares no table_name and no should_build_db, and extends no base model: a
// settings schema is metadata only, and the core.basemodel mixins would inject tenant_id and audit
// columns that a schema with no table must not carry.
//
// Neither field carries allow_override: true. These are the one case where the restrictive reading
// is right: a session timeout or an MFA requirement that an individual could opt out of is not a
// policy, it is a suggestion. The false is stated explicitly rather than omitted, because an
// absent allow_override reads as overridable.
//
// require_mfa has no default. A tenant that has not answered this question has not opted out of
// MFA; defaulting to false would silently record that it had.
func TenantSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(tenantSettingsSchemaJson)
}
