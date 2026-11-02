// Package settingsschema declares the layers of the settings_schema resource: one row per module
// and level, holding the setting definitions a module registered at boot.
//
// Over REST the resource is read-only. Rows are written by boot-time registration through the
// repository, never by a client.
package settingsschema

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type SettingsSchemaRepository interface {
	dynamicresource.CrudRepository
}

type SettingsSchemaDomainService interface {
	dynamicresource.CrudDomainService
}

type SettingsSchemaApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	GetSettingsSchemaByIdQuery = dynamicresource.GetByIdQuery
	SearchSettingsSchemasQuery = dynamicresource.SearchQuery
	SettingsSchemaExistsQuery  = dynamicresource.ExistsQuery
)

type (
	GetSettingsSchemaByIdResult = dynamicresource.GetOneResult
	SearchSettingsSchemasResult = dynamicresource.SearchResult
	SettingsSchemaExistsResult  = dynamicresource.ExistsResult
)
