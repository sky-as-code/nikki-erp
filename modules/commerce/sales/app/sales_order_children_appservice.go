package app

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The order's child resources.
//
// The line carries the quantity and snapshot rules in its domain service, so it keeps full CRUD.
// The other three are read-only: a component is written by repricing, an adjustment by a discount,
// an event by every transition, so a client that could write them directly could contradict the
// order they describe. CrudActions on their onions enforces that at the application layer, which
// holds for an in-process caller too, not just for a route.

func NewSalesOrderLineApplicationService(
	base dynamicresource.CrudApplicationService,
) itOrder.SalesOrderLineApplicationService {
	return &SalesOrderLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderLineApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewSalesOrderLineComponentApplicationService(base dynamicresource.CrudApplicationService) itOrder.SalesOrderLineComponentApplicationService {
	return &SalesOrderLineComponentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderLineComponentApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewSalesOrderAdjustmentApplicationService(base dynamicresource.CrudApplicationService) itOrder.SalesOrderAdjustmentApplicationService {
	return &SalesOrderAdjustmentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderAdjustmentApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewSalesOrderEventApplicationService(
	base dynamicresource.CrudApplicationService,
) itOrder.SalesOrderEventApplicationService {
	return &SalesOrderEventApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderEventApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
