package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external/invoicing"
	itReturns "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/returns"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The return resources. The return itself is writable and carries the three lifecycle actions;
// its lines and refund legs are records of what happened, so they are read-only.

type salesReturnEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_return"`
}

func registerSalesReturnEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesReturnSchemaName),
		func(
			param dynamicresource.BuildParam,
			dLock distributedlock.DistributedLock,
			settings itExt.EffectiveSettingsExtService,
			reservations itExt.FulfillmentReservationExtService,
			orders itExt.PaymentOrderExtService,
			fulfillment itExt.FulfillmentExtService,
			invoicing itInvoicing.InvoicingExtService,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesReturnSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesReturnRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesReturnApplicationService(
						base, dLock, settings, reservations, orders, fulfillment, invoicing)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesReturnEngineParam) itReturns.SalesReturnRepository {
			return p.Engine.Repository().(itReturns.SalesReturnRepository)
		},
		func(p salesReturnEngineParam) itReturns.SalesReturnApplicationService {
			return p.Engine.ApplicationService().(itReturns.SalesReturnApplicationService)
		},
	))
}

type salesReturnLineEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_return_line"`
}

func registerSalesReturnLineEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesReturnLineSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesReturnLineSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesReturnLineRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesReturnLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesReturnLineEngineParam) itReturns.SalesReturnLineRepository {
			return p.Engine.Repository().(itReturns.SalesReturnLineRepository)
		},
		func(p salesReturnLineEngineParam) itReturns.SalesReturnLineApplicationService {
			return p.Engine.ApplicationService().(itReturns.SalesReturnLineApplicationService)
		},
	))
}

type salesRefundPaymentEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_refund_payment"`
}

func registerSalesRefundPaymentEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesRefundPaymentSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesRefundPaymentSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesRefundPaymentRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesRefundPaymentApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesRefundPaymentEngineParam) itReturns.SalesRefundPaymentRepository {
			return p.Engine.Repository().(itReturns.SalesRefundPaymentRepository)
		},
		func(p salesRefundPaymentEngineParam) itReturns.SalesRefundPaymentApplicationService {
			return p.Engine.ApplicationService().(itReturns.SalesRefundPaymentApplicationService)
		},
	))
}
