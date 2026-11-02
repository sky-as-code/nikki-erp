package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/fulfillment"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The fulfillment resources. The fulfillment itself is writable and carries the dispense loop;
// the four records it produces are read-only.

type salesOrderFulfillmentEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment"`
}

func registerSalesOrderFulfillmentEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesOrderFulfillmentSchemaName),
		func(
			param dynamicresource.BuildParam,
			dLock distributedlock.DistributedLock,
			settings itExt.EffectiveSettingsExtService,
			reservations itExt.FulfillmentReservationExtService,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesOrderFulfillmentSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesOrderFulfillmentRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesOrderFulfillmentApplicationService(base, dLock, settings, reservations)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesOrderFulfillmentEngineParam) itFulfillment.SalesOrderFulfillmentRepository {
			return p.Engine.Repository().(itFulfillment.SalesOrderFulfillmentRepository)
		},
		func(p salesOrderFulfillmentEngineParam) itFulfillment.SalesOrderFulfillmentApplicationService {
			return p.Engine.ApplicationService().(itFulfillment.SalesOrderFulfillmentApplicationService)
		},
	))
}

type salesOrderFulfillmentItemEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment_item"`
}

func registerSalesOrderFulfillmentItemEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesOrderFulfillmentItemSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderFulfillmentItemSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesOrderFulfillmentItemRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesOrderFulfillmentItemApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesOrderFulfillmentItemEngineParam) itFulfillment.SalesOrderFulfillmentItemRepository {
			return p.Engine.Repository().(itFulfillment.SalesOrderFulfillmentItemRepository)
		},
		func(p salesOrderFulfillmentItemEngineParam) itFulfillment.SalesOrderFulfillmentItemApplicationService {
			return p.Engine.ApplicationService().(itFulfillment.SalesOrderFulfillmentItemApplicationService)
		},
	))
}

type salesFulfillmentAttemptEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt"`
}

func registerSalesFulfillmentAttemptEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesFulfillmentAttemptSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentAttemptSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesFulfillmentAttemptRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesFulfillmentAttemptApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesFulfillmentAttemptEngineParam) itFulfillment.SalesFulfillmentAttemptRepository {
			return p.Engine.Repository().(itFulfillment.SalesFulfillmentAttemptRepository)
		},
		func(p salesFulfillmentAttemptEngineParam) itFulfillment.SalesFulfillmentAttemptApplicationService {
			return p.Engine.ApplicationService().(itFulfillment.SalesFulfillmentAttemptApplicationService)
		},
	))
}

type salesFulfillmentAttemptItemEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt_item"`
}

func registerSalesFulfillmentAttemptItemEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesFulfillmentAttemptItemSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentAttemptItemSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesFulfillmentAttemptItemRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesFulfillmentAttemptItemApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesFulfillmentAttemptItemEngineParam) itFulfillment.SalesFulfillmentAttemptItemRepository {
			return p.Engine.Repository().(itFulfillment.SalesFulfillmentAttemptItemRepository)
		},
		func(p salesFulfillmentAttemptItemEngineParam) itFulfillment.SalesFulfillmentAttemptItemApplicationService {
			return p.Engine.ApplicationService().(itFulfillment.SalesFulfillmentAttemptItemApplicationService)
		},
	))
}

type salesFulfillmentTargetChangeEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_target_change"`
}

func registerSalesFulfillmentTargetChangeEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesFulfillmentTargetChangeSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentTargetChangeSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesFulfillmentTargetChangeRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesFulfillmentTargetChangeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesFulfillmentTargetChangeEngineParam) itFulfillment.SalesFulfillmentTargetChangeRepository {
			return p.Engine.Repository().(itFulfillment.SalesFulfillmentTargetChangeRepository)
		},
		func(p salesFulfillmentTargetChangeEngineParam) itFulfillment.SalesFulfillmentTargetChangeApplicationService {
			return p.Engine.ApplicationService().(itFulfillment.SalesFulfillmentTargetChangeApplicationService)
		},
	))
}
