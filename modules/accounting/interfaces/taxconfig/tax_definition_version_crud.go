package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxDefinitionVersionRepository interface {
	composable.CrudRepository
}

type TaxDefinitionVersionDomainService interface {
	composable.CrudDomainService
}

type TaxDefinitionVersionApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxDefinitionVersionCommand = composable.CreateCommand
	UpdateTaxDefinitionVersionCommand = composable.UpdateCommand
	DeleteTaxDefinitionVersionCommand = composable.DeleteCommand
	SearchTaxDefinitionVersionQuery   = composable.SearchQuery
)
