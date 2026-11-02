package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

// NewSettingsSchemaRepository is handed the composable default by the settings_schema onion.
func NewSettingsSchemaRepository(base dynamicresource.CrudRepository) itSchema.SettingsSchemaRepository {
	return &SettingsSchemaRepositoryImpl{CrudRepository: base}
}

type SettingsSchemaRepositoryImpl struct {
	dynamicresource.CrudRepository
}
