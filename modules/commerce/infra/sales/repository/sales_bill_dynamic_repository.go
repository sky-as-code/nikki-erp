package repository

import (
	itBilling "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/billing"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesBillRepository(base dynamicresource.CrudRepository) itBilling.SalesBillRepository {
	return &SalesBillRepositoryImpl{CrudRepository: base}
}

type SalesBillRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesBillLineRepository(base dynamicresource.CrudRepository) itBilling.SalesBillLineRepository {
	return &SalesBillLineRepositoryImpl{CrudRepository: base}
}

type SalesBillLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesBillRelationRepository(base dynamicresource.CrudRepository) itBilling.SalesBillRelationRepository {
	return &SalesBillRelationRepositoryImpl{CrudRepository: base}
}

type SalesBillRelationRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPaymentRepository(base dynamicresource.CrudRepository) itBilling.SalesPaymentRepository {
	return &SalesPaymentRepositoryImpl{CrudRepository: base}
}

type SalesPaymentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesFulfillmentRequestRepository(base dynamicresource.CrudRepository) itBilling.SalesFulfillmentRequestRepository {
	return &SalesFulfillmentRequestRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentRequestRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesFulfillmentRequestLineRepository(base dynamicresource.CrudRepository) itBilling.SalesFulfillmentRequestLineRepository {
	return &SalesFulfillmentRequestLineRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentRequestLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}
