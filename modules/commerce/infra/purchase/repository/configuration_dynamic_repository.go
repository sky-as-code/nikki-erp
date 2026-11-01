package repository

import (
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/configuration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewConfigurationRepository(base dynamicresource.CrudRepository) itConfiguration.ConfigurationCrudRepository {
	return &ConfigurationRepositoryImpl{CrudRepository: base}
}

type ConfigurationRepositoryImpl struct {
	dynamicresource.CrudRepository
}
