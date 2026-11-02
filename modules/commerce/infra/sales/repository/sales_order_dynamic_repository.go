package repository

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesOrderRepository(base dynamicresource.CrudRepository) itOrder.SalesOrderRepository {
	return &SalesOrderRepositoryImpl{CrudRepository: base}
}

type SalesOrderRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderLineRepository(base dynamicresource.CrudRepository) itOrder.SalesOrderLineRepository {
	return &SalesOrderLineRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderLineAllocationRepository(
	base dynamicresource.CrudRepository,
) itOrder.SalesOrderLineAllocationRepository {
	return &SalesOrderLineAllocationRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineAllocationRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderLineComponentRepository(base dynamicresource.CrudRepository) itOrder.SalesOrderLineComponentRepository {
	return &SalesOrderLineComponentRepositoryImpl{CrudRepository: base}
}

type SalesOrderLineComponentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderAdjustmentRepository(base dynamicresource.CrudRepository) itOrder.SalesOrderAdjustmentRepository {
	return &SalesOrderAdjustmentRepositoryImpl{CrudRepository: base}
}

type SalesOrderAdjustmentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesOrderEventRepository(base dynamicresource.CrudRepository) itOrder.SalesOrderEventRepository {
	return &SalesOrderEventRepositoryImpl{CrudRepository: base}
}

type SalesOrderEventRepositoryImpl struct {
	dynamicresource.CrudRepository
}
