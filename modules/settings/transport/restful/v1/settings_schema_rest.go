package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SettingsSchemaEngineName is the container name of the onion, checked against
// composable.EngineDependencyName by a test.
const SettingsSchemaEngineName = "dynengine_settings_schema"

type settingsSchemaRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_settings_schema"`
}

func NewSettingsSchemaRest(params settingsSchemaRestParams) *SettingsSchemaRest {
	rest := &SettingsSchemaRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SettingsSchemaRest serves the read-only built-ins of the resource.
type SettingsSchemaRest struct {
	composable.CrudRestBase
}
