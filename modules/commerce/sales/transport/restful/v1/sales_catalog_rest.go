// Package v1 serves the Sales resources over HTTP. Each Rest struct embeds the composable CRUD
// base for the built-in routes and adds one method per custom action; the method is a single line,
// because composable.ServeAction owns the whole HTTP contract -- binding, the 400 for a business
// refusal, the 404 for a missing record, and the shape of a success.
package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

// The container names of the catalogue onions, declared once for the struct tags and checked
// against composable.EngineDependencyName by a test.
const (
	SalesFulfillmentMethodEngineName = "dynengine_sales_fulfillment_method"
	SalesChannelEngineName           = "dynengine_sales_channel"
	SalesPointEngineName             = "dynengine_sales_point"
	SalesPricelistEngineName         = "dynengine_sales_pricelist"
	SalesPricelistItemEngineName     = "dynengine_sales_pricelist_item"
	SalesComboEngineName             = "dynengine_sales_combo"
	SalesComboComponentEngineName    = "dynengine_sales_combo_component"
)

type salesFulfillmentMethodRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_method"`
}

func NewSalesFulfillmentMethodRest(params salesFulfillmentMethodRestParams) *SalesFulfillmentMethodRest {
	rest := &SalesFulfillmentMethodRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.methodSvc = params.Engine.ApplicationService().(itCatalog.SalesFulfillmentMethodApplicationService)
	return rest
}

type SalesFulfillmentMethodRest struct {
	composable.CrudRestBase
	methodSvc itCatalog.SalesFulfillmentMethodApplicationService
}

type salesChannelRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_channel"`
}

func NewSalesChannelRest(params salesChannelRestParams) *SalesChannelRest {
	rest := &SalesChannelRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.channelSvc = params.Engine.ApplicationService().(itCatalog.SalesChannelApplicationService)
	return rest
}

type SalesChannelRest struct {
	composable.CrudRestBase
	channelSvc itCatalog.SalesChannelApplicationService
}

// The channel keeps one service field: the payment-method handlers delegate through it rather
// than holding a second service, so the REST layer has a single collaborator per resource.

type salesPointRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_point"`
}

func NewSalesPointRest(params salesPointRestParams) *SalesPointRest {
	rest := &SalesPointRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.pointSvc = params.Engine.ApplicationService().(itCatalog.SalesPointApplicationService)
	return rest
}

type SalesPointRest struct {
	composable.CrudRestBase
	pointSvc itCatalog.SalesPointApplicationService
}

type salesPricelistRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_pricelist"`
}

func NewSalesPricelistRest(params salesPricelistRestParams) *SalesPricelistRest {
	rest := &SalesPricelistRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.pricelistSvc = params.Engine.ApplicationService().(itCatalog.SalesPricelistApplicationService)
	return rest
}

type SalesPricelistRest struct {
	composable.CrudRestBase
	pricelistSvc itCatalog.SalesPricelistApplicationService
}

type salesPricelistItemRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_pricelist_item"`
}

func NewSalesPricelistItemRest(params salesPricelistItemRestParams) *SalesPricelistItemRest {
	rest := &SalesPricelistItemRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPricelistItemRest struct {
	composable.CrudRestBase
}

type salesComboRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_combo"`
}

func NewSalesComboRest(params salesComboRestParams) *SalesComboRest {
	rest := &SalesComboRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesComboRest struct {
	composable.CrudRestBase
}

type salesComboComponentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_combo_component"`
}

func NewSalesComboComponentRest(params salesComboComponentRestParams) *SalesComboComponentRest {
	rest := &SalesComboComponentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesComboComponentRest struct {
	composable.CrudRestBase
}
