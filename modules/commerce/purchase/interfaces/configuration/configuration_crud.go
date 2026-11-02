// Package configuration declares the CRUD layers of the per-org purchase configuration.
package configuration

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ConfigurationCrudRepository reads and writes purchase configurations.
type ConfigurationCrudRepository interface {
	dynamicresource.CrudRepository
}

// ConfigurationCrudDomainService is the CRUD of the resource.
type ConfigurationCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// ConfigurationCrudApplicationService is the authorized surface the REST handler serves.
type ConfigurationCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateConfigurationCommand      = dynamicresource.CreateCommand
	UpdateConfigurationCommand      = dynamicresource.UpdateCommand
	DeleteConfigurationCommand      = dynamicresource.DeleteCommand
	SetConfigurationArchivedCommand = dynamicresource.SetArchivedCommand
	GetConfigurationByIdQuery       = dynamicresource.GetByIdQuery
	SearchConfigurationsQuery       = dynamicresource.SearchQuery
	ConfigurationExistsQuery        = dynamicresource.ExistsQuery
)
