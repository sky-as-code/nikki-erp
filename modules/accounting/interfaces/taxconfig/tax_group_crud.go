package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxGroupRepository interface {
	composable.CrudRepository
}

type TaxGroupDomainService interface {
	composable.CrudDomainService
}

type TaxGroupApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxGroupCommand = composable.CreateCommand
	UpdateTaxGroupCommand = composable.UpdateCommand
	DeleteTaxGroupCommand = composable.DeleteCommand
	SearchTaxGroupQuery   = composable.SearchQuery
)
