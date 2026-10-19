package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxJurisdictionRepository interface {
	composable.CrudRepository
}

type TaxJurisdictionDomainService interface {
	composable.CrudDomainService
}

type TaxJurisdictionApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxJurisdictionCommand = composable.CreateCommand
	UpdateTaxJurisdictionCommand = composable.UpdateCommand
	DeleteTaxJurisdictionCommand = composable.DeleteCommand
	SearchTaxJurisdictionQuery   = composable.SearchQuery
)
