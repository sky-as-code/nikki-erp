package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

// NewCurrencyRepository is handed the composable default by the currency onion. The type exists
// so a bespoke query has a home; it adds nothing today.
func NewCurrencyRepository(base dynamicresource.CrudRepository) itCurrency.CurrencyRepository {
	return &CurrencyRepositoryImpl{CrudRepository: base}
}

type CurrencyRepositoryImpl struct {
	dynamicresource.CrudRepository
}
