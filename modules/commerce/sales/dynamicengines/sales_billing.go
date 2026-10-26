package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The billing resources. The bill takes its four ports as dig arguments, which retires
// SetPaymentMethodPort, SetPaymentOrderPort and SetChannelPaymentService.
//
// SetChannelPaymentService existed because ChannelPaymentAppService is one of Sales' own
// application services, registered several steps after the external ports, so resolving it
// eagerly in InitExternal was a same-module cycle. dig resolves it lazily here instead, and the
// ordering constraint disappears with the setter.

type salesBillEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill"`
}

func registerSalesBillEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesBillSchemaName),
		func(
			param composable.BuildParam,
			settings itExt.EffectiveSettingsExtService,
			dLock distributedlock.DistributedLock,
			methods itExt.PaymentMethodExtService,
			orders itExt.PaymentOrderExtService,
			channels itChannel.ChannelPaymentAppService,
			pointPayments *services.PointPaymentDomainServiceImpl,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesBillSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesBillRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesBillApplicationService(
						base, settings, dLock, methods, orders, channels, pointPayments)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesBillEngineParam) itBilling.SalesBillRepository {
			return p.Engine.Repository().(itBilling.SalesBillRepository)
		},
		func(p salesBillEngineParam) itBilling.SalesBillApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesBillApplicationService)
		},
	))
	return err
}

type salesBillLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill_line"`
}

func registerSalesBillLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesBillLineSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesBillLineSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesBillLineRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesBillLineApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesBillLineEngineParam) itBilling.SalesBillLineRepository {
			return p.Engine.Repository().(itBilling.SalesBillLineRepository)
		},
		func(p salesBillLineEngineParam) itBilling.SalesBillLineApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesBillLineApplicationService)
		},
	))
	return err
}

type salesBillRelationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill_relation"`
}

func registerSalesBillRelationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesBillRelationSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesBillRelationSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesBillRelationRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesBillRelationApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesBillRelationEngineParam) itBilling.SalesBillRelationRepository {
			return p.Engine.Repository().(itBilling.SalesBillRelationRepository)
		},
		func(p salesBillRelationEngineParam) itBilling.SalesBillRelationApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesBillRelationApplicationService)
		},
	))
	return err
}

type salesPaymentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_payment"`
}

func registerSalesPaymentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPaymentSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesPaymentSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPaymentRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPaymentApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesPaymentEngineParam) itBilling.SalesPaymentRepository {
			return p.Engine.Repository().(itBilling.SalesPaymentRepository)
		},
		func(p salesPaymentEngineParam) itBilling.SalesPaymentApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesPaymentApplicationService)
		},
	))
	return err
}

type salesFulfillmentRequestEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request"`
}

func registerSalesFulfillmentRequestEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentRequestSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentRequestSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentRequestRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesFulfillmentRequestApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesFulfillmentRequestEngineParam) itBilling.SalesFulfillmentRequestRepository {
			return p.Engine.Repository().(itBilling.SalesFulfillmentRequestRepository)
		},
		func(p salesFulfillmentRequestEngineParam) itBilling.SalesFulfillmentRequestApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesFulfillmentRequestApplicationService)
		},
	))
	return err
}

type salesFulfillmentRequestLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request_line"`
}

func registerSalesFulfillmentRequestLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentRequestLineSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentRequestLineSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentRequestLineRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesFulfillmentRequestLineApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesFulfillmentRequestLineEngineParam) itBilling.SalesFulfillmentRequestLineRepository {
			return p.Engine.Repository().(itBilling.SalesFulfillmentRequestLineRepository)
		},
		func(p salesFulfillmentRequestLineEngineParam) itBilling.SalesFulfillmentRequestLineApplicationService {
			return p.Engine.ApplicationService().(itBilling.SalesFulfillmentRequestLineApplicationService)
		},
	))
	return err
}
