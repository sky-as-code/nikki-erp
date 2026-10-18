package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SettingsRecordEngineName is the container name of the onion, checked against
// composable.EngineDependencyName by a test.
const SettingsRecordEngineName = "dynengine_settings_record"

type settingsRecordRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_settings_record"`
}

func NewSettingsRecordRest(params settingsRecordRestParams) *SettingsRecordRest {
	rest := &SettingsRecordRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SettingsRecordRest serves the read-only built-ins of the resource.
type SettingsRecordRest struct {
	composable.CrudRestBase
}
