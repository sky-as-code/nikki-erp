package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxGroupRepository interface {
	dynamicresource.CrudRepository
}

type TaxGroupDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxGroupApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxGroupCommand = dynamicresource.CreateCommand
	UpdateTaxGroupCommand = dynamicresource.UpdateCommand
	DeleteTaxGroupCommand = dynamicresource.DeleteCommand
	SearchTaxGroupQuery   = dynamicresource.SearchQuery
)
