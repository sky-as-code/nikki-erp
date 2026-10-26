package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
)

func NewSalesBillRepository(base composable.CrudRepository) itBilling.SalesBillRepository {
	return &SalesBillRepositoryImpl{CrudRepository: base}
}

type SalesBillRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesBillLineRepository(base composable.CrudRepository) itBilling.SalesBillLineRepository {
	return &SalesBillLineRepositoryImpl{CrudRepository: base}
}

type SalesBillLineRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesBillRelationRepository(base composable.CrudRepository) itBilling.SalesBillRelationRepository {
	return &SalesBillRelationRepositoryImpl{CrudRepository: base}
}

type SalesBillRelationRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPaymentRepository(base composable.CrudRepository) itBilling.SalesPaymentRepository {
	return &SalesPaymentRepositoryImpl{CrudRepository: base}
}

type SalesPaymentRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesFulfillmentRequestRepository(base composable.CrudRepository) itBilling.SalesFulfillmentRequestRepository {
	return &SalesFulfillmentRequestRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentRequestRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesFulfillmentRequestLineRepository(base composable.CrudRepository) itBilling.SalesFulfillmentRequestLineRepository {
	return &SalesFulfillmentRequestLineRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentRequestLineRepositoryImpl struct {
	composable.CrudRepository
}
