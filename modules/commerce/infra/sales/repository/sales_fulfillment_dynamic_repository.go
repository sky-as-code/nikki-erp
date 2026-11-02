package repository

import (
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/fulfillment"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesOrderFulfillmentRepository(base dynamicresource.CrudRepository) itFulfillment.SalesOrderFulfillmentRepository {
	return &SalesOrderFulfillmentRepositoryImpl{CrudRepository: base}
}

type SalesOrderFulfillmentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderFulfillmentItemRepository(base dynamicresource.CrudRepository) itFulfillment.SalesOrderFulfillmentItemRepository {
	return &SalesOrderFulfillmentItemRepositoryImpl{CrudRepository: base}
}

type SalesOrderFulfillmentItemRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesFulfillmentAttemptRepository(
	base dynamicresource.CrudRepository,
) itFulfillment.SalesFulfillmentAttemptRepository {
	return &SalesFulfillmentAttemptRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentAttemptRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesFulfillmentAttemptItemRepository(base dynamicresource.CrudRepository) itFulfillment.SalesFulfillmentAttemptItemRepository {
	return &SalesFulfillmentAttemptItemRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentAttemptItemRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesFulfillmentTargetChangeRepository(base dynamicresource.CrudRepository) itFulfillment.SalesFulfillmentTargetChangeRepository {
	return &SalesFulfillmentTargetChangeRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentTargetChangeRepositoryImpl struct {
	dynamicresource.CrudRepository
}
