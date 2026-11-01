package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

// NewSettingsSchemaApplicationService is handed the composable default by the onion: the authorized reads.
func NewSettingsSchemaApplicationService(
	base dynamicresource.CrudApplicationService,
) itSchema.SettingsSchemaApplicationService {
	return &SettingsSchemaApplicationServiceImpl{CrudApplicationService: base}
}

type SettingsSchemaApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
