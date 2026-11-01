package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SettingsQuotaEngineName is the container name of the onion, checked against
// dynamicresource.EngineDependencyName by a test.
const SettingsQuotaEngineName = "dynengine_settings_quota"

type settingsQuotaRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_settings_quota"`
}

func NewSettingsQuotaRest(params settingsQuotaRestParams) *SettingsQuotaRest {
	rest := &SettingsQuotaRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// SettingsQuotaRest serves the read-only built-ins of the resource.
type SettingsQuotaRest struct {
	dynamicresource.CrudRestBase
}
