package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
)

func NewSalesOrderRepository(base composable.CrudRepository) itOrder.SalesOrderRepository {
	return &SalesOrderRepositoryImpl{CrudRepository: base}
}

type SalesOrderRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderLineRepository(base composable.CrudRepository) itOrder.SalesOrderLineRepository {
	return &SalesOrderLineRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderLineAllocationRepository(
	base composable.CrudRepository,
) itOrder.SalesOrderLineAllocationRepository {
	return &SalesOrderLineAllocationRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineAllocationRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderLineComponentRepository(base composable.CrudRepository) itOrder.SalesOrderLineComponentRepository {
	return &SalesOrderLineComponentRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineComponentRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderAdjustmentRepository(base composable.CrudRepository) itOrder.SalesOrderAdjustmentRepository {
	return &SalesOrderAdjustmentRepositoryImpl{CrudRepository: base}
}

type SalesOrderAdjustmentRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesOrderEventRepository(base composable.CrudRepository) itOrder.SalesOrderEventRepository {
	return &SalesOrderEventRepositoryImpl{CrudRepository: base}
}

type SalesOrderEventRepositoryImpl struct {
	composable.CrudRepository
}
