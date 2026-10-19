package taxconfig

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TaxProductClassificationRepository interface {
	composable.CrudRepository
}

type TaxProductClassificationDomainService interface {
	composable.CrudDomainService
}

type TaxProductClassificationApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTaxProductClassificationCommand = composable.CreateCommand
	UpdateTaxProductClassificationCommand = composable.UpdateCommand
	DeleteTaxProductClassificationCommand = composable.DeleteCommand
	SearchTaxProductClassificationQuery   = composable.SearchQuery
)
