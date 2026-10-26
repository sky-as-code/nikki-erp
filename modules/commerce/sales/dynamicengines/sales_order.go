package dynamicengines

import (
	stdErr "errors"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itOrder "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
)

// The order resources.
//
// The order's closure takes its ports as dig arguments rather than reading them from package
// variables. That is the substance of this migration for this resource: a legacy action callback
// was handed nothing, so seven ports had to be pushed into the package before the first request
// arrived, and Init's step order was load-bearing because of it. Here dig resolves them, so the
// ordering constraint is gone.
//
// Two ports are optional because no adapter ships for them in every deployment. They must stay
// optional: the forcing Invoke below builds this onion at boot, so a required port with no
// provider would fail the boot rather than the first request that needs it.

type salesOrderEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order"`
}

func registerSalesOrderEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderSchemaName),
		func(
			param composable.BuildParam,
			tax itExt.TaxCalculationExtService,
			settings itExt.EffectiveSettingsExtService,
			dLock distributedlock.DistributedLock,
			products itExt.ProductVariantExtService,
			fulfillment itExt.FulfillmentExtService,
			basis itExt.ProductPricingBasisExtService,
			parties itExt.PartyExtService,
			reservations itExt.FulfillmentReservationExtService,
			paymentOrders itExt.PaymentOrderExtService,
			invoicing itInvoicing.InvoicingExtService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesOrderSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesOrderDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesOrderCrudApplicationService(
						base, tax, settings, dLock, products, fulfillment, basis, reservations, parties,
						paymentOrders, invoicing)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesOrderEngineParam) itOrder.SalesOrderRepository {
			return p.Engine.Repository().(itOrder.SalesOrderRepository)
		},
		func(p salesOrderEngineParam) itOrder.SalesOrderDomainService {
			return p.Engine.DomainService().(itOrder.SalesOrderDomainService)
		},
		func(p salesOrderEngineParam) itOrder.SalesOrderApplicationService {
			return p.Engine.ApplicationService().(itOrder.SalesOrderApplicationService)
		},
	))
	return err
}

type salesOrderLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_line"`
}

func registerSalesOrderLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderLineSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesOrderLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderLineRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesOrderLineDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesOrderLineApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesOrderLineEngineParam) itOrder.SalesOrderLineRepository {
			return p.Engine.Repository().(itOrder.SalesOrderLineRepository)
		},
		func(p salesOrderLineEngineParam) itOrder.SalesOrderLineDomainService {
			return p.Engine.DomainService().(itOrder.SalesOrderLineDomainService)
		},
		func(p salesOrderLineEngineParam) itOrder.SalesOrderLineApplicationService {
			return p.Engine.ApplicationService().(itOrder.SalesOrderLineApplicationService)
		},
	))
	return err
}

// Allocations are an internal child of an order line. The create-order service writes them in the
// same transaction as their line, so this engine exists to provide its repository and domain
// service, not to expose a client CRUD surface.
type salesOrderLineAllocationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_line_allocation"`
}

func registerSalesOrderLineAllocationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderLineAllocationSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderLineAllocationSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderLineAllocationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesOrderLineAllocationDomainService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesOrderLineAllocationEngineParam) itOrder.SalesOrderLineAllocationRepository {
			return p.Engine.Repository().(itOrder.SalesOrderLineAllocationRepository)
		},
		func(p salesOrderLineAllocationEngineParam) itOrder.SalesOrderLineAllocationDomainService {
			return p.Engine.DomainService().(itOrder.SalesOrderLineAllocationDomainService)
		},
	))
}

// readOnlyCrudActions is the allow-list for the resources an order writes about itself. It is
// checked in the application layer, so it holds for an in-process caller as well as for a route:
// a component, an adjustment or an event that a client could write directly could contradict the
// order it describes.
func readOnlyCrudActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionGetById,
		composable.CrudActionSearch,
		composable.CrudActionExists,
		composable.CrudActionGetSchema,
	}
}

type salesOrderLineComponentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_line_component"`
}

func registerSalesOrderLineComponentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderLineComponentSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderLineComponentSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderLineComponentRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesOrderLineComponentApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesOrderLineComponentEngineParam) itOrder.SalesOrderLineComponentRepository {
			return p.Engine.Repository().(itOrder.SalesOrderLineComponentRepository)
		},
		func(p salesOrderLineComponentEngineParam) itOrder.SalesOrderLineComponentApplicationService {
			return p.Engine.ApplicationService().(itOrder.SalesOrderLineComponentApplicationService)
		},
	))
	return err
}

type salesOrderAdjustmentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_adjustment"`
}

func registerSalesOrderAdjustmentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderAdjustmentSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderAdjustmentSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderAdjustmentRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesOrderAdjustmentApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesOrderAdjustmentEngineParam) itOrder.SalesOrderAdjustmentRepository {
			return p.Engine.Repository().(itOrder.SalesOrderAdjustmentRepository)
		},
		func(p salesOrderAdjustmentEngineParam) itOrder.SalesOrderAdjustmentApplicationService {
			return p.Engine.ApplicationService().(itOrder.SalesOrderAdjustmentApplicationService)
		},
	))
	return err
}

type salesOrderEventEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_event"`
}

func registerSalesOrderEventEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderEventSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderEventSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderEventRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesOrderEventApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesOrderEventEngineParam) itOrder.SalesOrderEventRepository {
			return p.Engine.Repository().(itOrder.SalesOrderEventRepository)
		},
		func(p salesOrderEventEngineParam) itOrder.SalesOrderEventApplicationService {
			return p.Engine.ApplicationService().(itOrder.SalesOrderEventApplicationService)
		},
	))
	return err
}
