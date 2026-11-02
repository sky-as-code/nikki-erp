package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TaxProductClassificationRepository interface {
	dynamicresource.CrudRepository
}

type TaxProductClassificationDomainService interface {
	dynamicresource.CrudDomainService
}

type TaxProductClassificationApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTaxProductClassificationCommand = dynamicresource.CreateCommand
	UpdateTaxProductClassificationCommand = dynamicresource.UpdateCommand
	DeleteTaxProductClassificationCommand = dynamicresource.DeleteCommand
	SearchTaxProductClassificationQuery   = dynamicresource.SearchQuery
)
