package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

func NewTransactionRepository(base dynamicresource.CrudRepository) itTransaction.TransactionCrudRepository {
	return &TransactionRepositoryImpl{CrudRepository: base}
}

type TransactionRepositoryImpl struct {
	dynamicresource.CrudRepository
}
