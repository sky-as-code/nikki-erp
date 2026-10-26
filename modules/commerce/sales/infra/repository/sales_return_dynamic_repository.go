package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itReturns "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/returns"
)

func NewSalesReturnRepository(base composable.CrudRepository) itReturns.SalesReturnRepository {
	return &SalesReturnRepositoryImpl{CrudRepository: base}
}

type SalesReturnRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesReturnLineRepository(base composable.CrudRepository) itReturns.SalesReturnLineRepository {
	return &SalesReturnLineRepositoryImpl{CrudRepository: base}
}

type SalesReturnLineRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesRefundPaymentRepository(base composable.CrudRepository) itReturns.SalesRefundPaymentRepository {
	return &SalesRefundPaymentRepositoryImpl{CrudRepository: base}
}

type SalesRefundPaymentRepositoryImpl struct {
	composable.CrudRepository
}
