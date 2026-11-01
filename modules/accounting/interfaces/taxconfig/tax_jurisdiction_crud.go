package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxJurisdictionRepository interface {
	dynamicresource.CrudRepository
}

type TaxJurisdictionDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxJurisdictionApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxJurisdictionCommand = dynamicresource.CreateCommand
	UpdateTaxJurisdictionCommand = dynamicresource.UpdateCommand
	DeleteTaxJurisdictionCommand = dynamicresource.DeleteCommand
	SearchTaxJurisdictionQuery   = dynamicresource.SearchQuery
)
