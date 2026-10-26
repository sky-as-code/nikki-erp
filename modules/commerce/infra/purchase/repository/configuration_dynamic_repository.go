package repository

import (
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/configuration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewConfigurationRepository(base composable.CrudRepository) itConfiguration.ConfigurationCrudRepository {
	return &ConfigurationRepositoryImpl{CrudRepository: base}
}

type ConfigurationRepositoryImpl struct {
	composable.CrudRepository
}
