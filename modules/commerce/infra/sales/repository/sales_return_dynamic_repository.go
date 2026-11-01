package repository

import (
	itReturns "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/returns"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesReturnRepository(base dynamicresource.CrudRepository) itReturns.SalesReturnRepository {
	return &SalesReturnRepositoryImpl{CrudRepository: base}
}

type SalesReturnRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesReturnLineRepository(base dynamicresource.CrudRepository) itReturns.SalesReturnLineRepository {
	return &SalesReturnLineRepositoryImpl{CrudRepository: base}
}

type SalesReturnLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesRefundPaymentRepository(base dynamicresource.CrudRepository) itReturns.SalesRefundPaymentRepository {
	return &SalesRefundPaymentRepositoryImpl{CrudRepository: base}
}

type SalesRefundPaymentRepositoryImpl struct {
	dynamicresource.CrudRepository
}
