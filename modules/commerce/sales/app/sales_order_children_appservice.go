package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
)

// The order's child resources.
//
// The line carries the quantity and snapshot rules in its domain service, so it keeps full CRUD.
// The other three are read-only: a component is written by repricing, an adjustment by a discount,
// an event by every transition, so a client that could write them directly could contradict the
// order they describe. CrudActions on their onions enforces that at the application layer, which
// holds for an in-process caller too, not just for a route.

func NewSalesOrderLineApplicationService(base composable.CrudApplicationService) itOrder.SalesOrderLineApplicationService {
	return &SalesOrderLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesOrderLineComponentApplicationService(base composable.CrudApplicationService) itOrder.SalesOrderLineComponentApplicationService {
	return &SalesOrderLineComponentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderLineComponentApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesOrderAdjustmentApplicationService(base composable.CrudApplicationService) itOrder.SalesOrderAdjustmentApplicationService {
	return &SalesOrderAdjustmentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderAdjustmentApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesOrderEventApplicationService(base composable.CrudApplicationService) itOrder.SalesOrderEventApplicationService {
	return &SalesOrderEventApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderEventApplicationServiceImpl struct {
	composable.CrudApplicationService
}
