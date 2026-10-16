// Package settings is the shared contract every feature module writes its settings through.
//
// A module owns the update endpoint for its own settings: it validates the items against its own
// schema and its own data, then persists them through SettingsService. The Settings module
// implements the service, but core cannot name it (settings depends on core), so apptrait
// registers the implementation at runtime -- the same seam requestguard uses.
package settings

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// Setting levels. A level says which kind of owner a setting belongs to.
const (
	LevelTenant = "tenant"
	LevelOrg    = "org"
	LevelUser   = "user"
)

// IsKnownLevel reports whether level names one of the three setting levels.
func IsKnownLevel(level string) bool {
	return level == LevelTenant || level == LevelOrg || level == LevelUser
}

// SettingsService persists a module's settings at one level, for the owner the acting user stands
// for at that level.
//
// The implementation authorizes exactly as the Settings module's application services do: a
// tenant or org write requires update on settings_record, a user write is the actor's own.
type SettingsService interface {
	SetSettings(ctx corectx.Context, level string, cmd SetSettingsCommand) (*SetSettingsResult, error)
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

type SetSettingsResult = dyn.OpResult[SetSettingsResultData]
