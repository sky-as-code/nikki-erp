package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fulfillment"
)

func NewSalesOrderFulfillmentRepository(base composable.CrudRepository) itFulfillment.SalesOrderFulfillmentRepository {
	return &SalesOrderFulfillmentRepositoryImpl{CrudRepository: base}
}

type SalesOrderFulfillmentRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderFulfillmentItemRepository(base composable.CrudRepository) itFulfillment.SalesOrderFulfillmentItemRepository {
	return &SalesOrderFulfillmentItemRepositoryImpl{CrudRepository: base}
}

type SalesOrderFulfillmentItemRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesFulfillmentAttemptRepository(base composable.CrudRepository) itFulfillment.SalesFulfillmentAttemptRepository {
	return &SalesFulfillmentAttemptRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentAttemptRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesFulfillmentAttemptItemRepository(base composable.CrudRepository) itFulfillment.SalesFulfillmentAttemptItemRepository {
	return &SalesFulfillmentAttemptItemRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentAttemptItemRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesFulfillmentTargetChangeRepository(base composable.CrudRepository) itFulfillment.SalesFulfillmentTargetChangeRepository {
	return &SalesFulfillmentTargetChangeRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentTargetChangeRepositoryImpl struct {
	composable.CrudRepository
}
