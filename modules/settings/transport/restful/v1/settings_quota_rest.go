package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SettingsQuotaEngineName is the container name of the onion, checked against
// composable.EngineDependencyName by a test.
const SettingsQuotaEngineName = "dynengine_settings_quota"

type settingsQuotaRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_settings_quota"`
}

func NewSettingsQuotaRest(params settingsQuotaRestParams) *SettingsQuotaRest {
	rest := &SettingsQuotaRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SettingsQuotaRest serves the read-only built-ins of the resource.
type SettingsQuotaRest struct {
	composable.CrudRestBase
}
