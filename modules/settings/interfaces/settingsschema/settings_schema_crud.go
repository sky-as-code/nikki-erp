// Package settingsschema declares the layers of the settings_schema resource: one row per module
// and level, holding the setting definitions a module registered at boot.
//
// Over REST the resource is read-only. Rows are written by boot-time registration through the
// repository, never by a client.
package settingsschema

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type SettingsSchemaRepository interface {
	composable.CrudRepository
}

type SettingsSchemaDomainService interface {
	composable.CrudDomainService
}

type SettingsSchemaApplicationService interface {
	composable.CrudApplicationService
}

type (
	GetSettingsSchemaByIdQuery = composable.GetByIdQuery
	SearchSettingsSchemasQuery = composable.SearchQuery
	SettingsSchemaExistsQuery  = composable.ExistsQuery
)

type (
	GetSettingsSchemaByIdResult = composable.GetOneResult
	SearchSettingsSchemasResult = composable.SearchResult
	SettingsSchemaExistsResult  = composable.ExistsResult
)
