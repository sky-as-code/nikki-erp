// Package dynamicengines declares the resource onions the Settings module serves through the
// composable resource engine, and registers them into the dependency container during the
// module's Init().
//
// Both resources are read-only over REST: rows are written by boot-time registration, by the
// per-level write path of each module (core/settings port) and by owner initialization, all of
// which go through the repositories published here. The behaviour that makes settings more than a
// key-value table (the partial save, the level authorization and the enforcement fan-out) lives in
// SettingsDomainServiceImpl, because it spans many rows of one transaction.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// InitDynamicEngines registers every resource onion this module owns.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerSettingsSchemaEngine(),
		registerSettingsRecordEngine(),
		registerSettingsQuotaEngine(),
	)
}

// ReadOnlyCrudActions is the built-in action set both settings resources serve. It is also the
// onion's allow-list, so a write is refused at the application layer as well as having no route.
func ReadOnlyCrudActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionGetById, composable.CrudActionGetByUnique, composable.CrudActionSearch,
		composable.CrudActionExists, composable.CrudActionGetSchema, composable.CrudActionComputeField,
	}
}
