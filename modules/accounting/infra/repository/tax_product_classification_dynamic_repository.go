package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxProductClassificationRepository is handed the composable default by the onion.
func NewTaxProductClassificationRepository(
	base dynamicresource.CrudRepository,
) taxconfig.TaxProductClassificationRepository {
	return &TaxProductClassificationRepositoryImpl{CrudRepository: base}
}

type TaxProductClassificationRepositoryImpl struct {
	dynamicresource.CrudRepository
}
