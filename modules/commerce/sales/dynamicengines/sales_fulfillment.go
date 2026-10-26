package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fulfillment"
)

// The fulfillment resources. The fulfillment itself is writable and carries the dispense loop;
// the four records it produces are read-only.

type salesOrderFulfillmentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment"`
}

func registerSalesOrderFulfillmentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderFulfillmentSchemaName),
		func(
			param composable.BuildParam,
			dLock distributedlock.DistributedLock,
			settings itExt.EffectiveSettingsExtService,
			reservations itExt.FulfillmentReservationExtService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesOrderFulfillmentSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderFulfillmentRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment_item"`
}

func registerSalesOrderFulfillmentItemEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesOrderFulfillmentItemSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesOrderFulfillmentItemSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesOrderFulfillmentItemRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt"`
}

func registerSalesFulfillmentAttemptEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentAttemptSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentAttemptSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentAttemptRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt_item"`
}

func registerSalesFulfillmentAttemptItemEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentAttemptItemSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentAttemptItemSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentAttemptItemRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_target_change"`
}

func registerSalesFulfillmentTargetChangeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentTargetChangeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesFulfillmentTargetChangeSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentTargetChangeRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
