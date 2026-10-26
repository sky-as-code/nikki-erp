package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/configuration"
)

func NewConfigurationRepository(base composable.CrudRepository) itConfiguration.ConfigurationCrudRepository {
	return &ConfigurationRepositoryImpl{CrudRepository: base}
}

type ConfigurationRepositoryImpl struct {
	composable.CrudRepository
}
