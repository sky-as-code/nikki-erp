// Package settings declares the capabilities the Settings module exposes to the rest of the
// application: schema registration, and the per-level read and write of setting values.
//
// It holds contracts and data only. Implementations live in the module's domain/services.
package settings

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// RegisterSchemaCommand asks Settings to record one module's setting definitions for one level.
//
// It is idempotent by [ModuleKey, Level]: re-registering an identical schema is a successful
// no-op, which is what lets it run on every boot.
type RegisterSchemaCommand struct {
	ModuleKey string
	Level     string

	// Schema defines the setting names, their types, defaults and per-field allow_override
	// metadata. It must be metadata-only: a settings schema describes values, it does not own a
	// table, and the only tables this module builds are its own two.
	Schema *dmodel.ModelSchema
}

type RegisterSchemaResultData struct {
	Id model.Id
	// Created is false when an identical registration already existed.
	Created bool
}

// SettingItem is one setting as the UI and consuming modules see it: the stored value together
// with enough of its declaration to render and validate it.
type SettingItem struct {
	Name  string
	Level string
	Value any

	// HasValue is false when no row exists for this owner yet and Value came from the schema's
	// declared default. The distinction matters on write: a default is not a stored choice.
	HasValue bool

	// AllowOverride is whether an owner below the tenant may keep its own value. Stored on the
	// row so a tenant admin decides it per tenant; a row nobody has ruled on falls back to the
	// module's declared metadata. False means the tenant's value has been written onto this row.
	AllowOverride bool

	// Editable is false when this actor may not change the item, either because the level is not
	// theirs or because AllowOverride is false and they are below the tenant.
	Editable bool

	// Field is the declaration the value was validated against, so a caller can render the right
	// control without re-reading the schema.
	Field *dmodel.ModelField
}

// GetSettingsQuery reads the settings one actor may see for one module.
type GetSettingsQuery struct {
	ModuleKey string
}

type GetSettingsResultData struct {
	ModuleKey string

	// OwnerType is the scope this actor is reading as: tenant, org or user.
	//
	// Reported because `Editable` alone cannot tell a client *who* it is talking to. A tenant
	// admin and an ordinary user both see editable items at their own level, so a pane deciding
	// whether to offer tenant-only controls -- changing the override policy, chiefly -- has no
	// other way to know. On the envelope rather than on each item: it describes the caller, not
	// the setting, and repeating it per item would invite the two to disagree.
	OwnerType string

	// Items are grouped by level for a tenant admin, and hold a single level for anyone else.
	Items []SettingItem
}

// SetSettingsCommand writes a partial set of changed items for one owner at one level.
//
// Only the items the caller actually changed belong here. An absent item is left untouched rather
// than cleared, and — because writes are last-write-wins with no version check — submitting an
// unchanged item is how a concurrent edit to a field this caller never touched gets clobbered.
type SetSettingsCommand struct {
	ModuleKey string
	Items     []SetSettingItem
}

type SetSettingItem struct {
	Name  string
	Value any

	// AllowOverride, when set, changes whether owners below the tenant may keep their own value.
	//
	// Only a tenant owner may send it: the flag is what locks the levels below, so an org or user
	// setting it would be escalating their own privilege. Nil leaves the stored policy alone,
	// which is what every non-tenant caller sends.
	AllowOverride *bool
}

type SetSettingsResultData struct {
	// Updated counts the rows written, including the children an enforced tenant setting fanned
	// out onto.
	Updated int
}

// InitOwnerSettingsCommand seeds a newly created organization or user with its own copy of the
// tenant's settings rows.
type InitOwnerSettingsCommand struct {
	TenantId  model.Id
	OwnerType string
	OwnerId   model.Id
}

type InitOwnerSettingsResultData struct {
	Created int
}
