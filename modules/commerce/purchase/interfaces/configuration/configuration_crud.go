// Package configuration declares the CRUD layers of the per-org purchase configuration.
package configuration

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ConfigurationCrudRepository reads and writes purchase configurations.
type ConfigurationCrudRepository interface {
	composable.CrudRepository
}

// ConfigurationCrudDomainService is the CRUD of the resource.
type ConfigurationCrudDomainService interface {
	composable.CrudDomainService
}

// ConfigurationCrudApplicationService is the authorized surface the REST handler serves.
type ConfigurationCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateConfigurationCommand      = composable.CreateCommand
	UpdateConfigurationCommand      = composable.UpdateCommand
	DeleteConfigurationCommand      = composable.DeleteCommand
	SetConfigurationArchivedCommand = composable.SetArchivedCommand
	GetConfigurationByIdQuery       = composable.GetByIdQuery
	SearchConfigurationsQuery       = composable.SearchQuery
	ConfigurationExistsQuery        = composable.ExistsQuery
)
