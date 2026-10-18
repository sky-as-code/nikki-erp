package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

func NewTransactionRepository(base composable.CrudRepository) itTransaction.TransactionCrudRepository {
	return &TransactionRepositoryImpl{CrudRepository: base}
}

type TransactionRepositoryImpl struct {
	composable.CrudRepository
}
