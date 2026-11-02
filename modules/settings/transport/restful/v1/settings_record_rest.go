package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SettingsRecordEngineName is the container name of the onion, checked against
// dynamicresource.EngineDependencyName by a test.
const SettingsRecordEngineName = "dynengine_settings_record"

type settingsRecordRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_settings_record"`
}

func NewSettingsRecordRest(params settingsRecordRestParams) *SettingsRecordRest {
	rest := &SettingsRecordRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SettingsRecordRest serves the read-only built-ins of the resource.
type SettingsRecordRest struct {
	dynamicresource.CrudRestBase
}
