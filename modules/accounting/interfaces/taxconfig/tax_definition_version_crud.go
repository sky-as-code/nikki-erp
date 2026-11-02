package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxDefinitionVersionRepository interface {
	dynamicresource.CrudRepository
}

type TaxDefinitionVersionDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxDefinitionVersionApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxDefinitionVersionCommand = dynamicresource.CreateCommand
	UpdateTaxDefinitionVersionCommand = dynamicresource.UpdateCommand
	DeleteTaxDefinitionVersionCommand = dynamicresource.DeleteCommand
	SearchTaxDefinitionVersionQuery   = dynamicresource.SearchQuery
)
