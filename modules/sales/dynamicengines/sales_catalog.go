package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
)

// The catalogue resources: the fulfillment policies, channels and points a sale happens through,
// and the pricelists and combos it is priced by.
//
// Each registration ends with a forcing Invoke. Registration alone builds nothing -- deps records
// the constructor and dig constructs on first resolution -- so without it a resource that no route
// and no sibling injects would never run InstallResource, and the first caller to reach it through
// the resource hub would fail at runtime, typically a cron sweep hours after a boot that looked
// healthy.

type salesFulfillmentMethodEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_method"`
}

func registerSalesFulfillmentMethodEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFulfillmentMethodSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesFulfillmentMethodSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFulfillmentMethodRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesFulfillmentMethodDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesFulfillmentMethodApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesFulfillmentMethodEngineParam) itCatalog.SalesFulfillmentMethodRepository {
			return p.Engine.Repository().(itCatalog.SalesFulfillmentMethodRepository)
		},
		func(p salesFulfillmentMethodEngineParam) itCatalog.SalesFulfillmentMethodDomainService {
			return p.Engine.DomainService().(itCatalog.SalesFulfillmentMethodDomainService)
		},
		func(p salesFulfillmentMethodEngineParam) itCatalog.SalesFulfillmentMethodApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesFulfillmentMethodApplicationService)
		},
	))
	return err
}

type salesChannelEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_channel"`
}

func registerSalesChannelEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesChannelSchemaName),
		func(
			param composable.BuildParam,
			channelPayments itChannel.ChannelPaymentAppService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesChannelSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesChannelRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesChannelDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesChannelCrudApplicationService(base, channelPayments)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesChannelEngineParam) itCatalog.SalesChannelRepository {
			return p.Engine.Repository().(itCatalog.SalesChannelRepository)
		},
		func(p salesChannelEngineParam) itCatalog.SalesChannelDomainService {
			return p.Engine.DomainService().(itCatalog.SalesChannelDomainService)
		},
		func(p salesChannelEngineParam) itCatalog.SalesChannelApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesChannelApplicationService)
		},
	))
	return err
}

type salesPointEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_point"`
}

func registerSalesPointEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPointSchemaName),
		func(
			param composable.BuildParam,
			pointPayments itChannel.PointPaymentAppService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPointSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPointRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesPointDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPointCrudApplicationService(base, pointPayments)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesPointEngineParam) itCatalog.SalesPointRepository {
			return p.Engine.Repository().(itCatalog.SalesPointRepository)
		},
		func(p salesPointEngineParam) itCatalog.SalesPointDomainService {
			return p.Engine.DomainService().(itCatalog.SalesPointDomainService)
		},
		func(p salesPointEngineParam) itCatalog.SalesPointApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesPointApplicationService)
		},
	))
	return err
}

type salesPricelistEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_pricelist"`
}

func registerSalesPricelistEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPricelistSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPricelistSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPricelistRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesPricelistDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPricelistApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesPricelistEngineParam) itCatalog.SalesPricelistRepository {
			return p.Engine.Repository().(itCatalog.SalesPricelistRepository)
		},
		func(p salesPricelistEngineParam) itCatalog.SalesPricelistDomainService {
			return p.Engine.DomainService().(itCatalog.SalesPricelistDomainService)
		},
		func(p salesPricelistEngineParam) itCatalog.SalesPricelistApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesPricelistApplicationService)
		},
	))
	return err
}

type salesPricelistItemEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_pricelist_item"`
}

func registerSalesPricelistItemEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPricelistItemSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPricelistItemSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPricelistItemRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSalesPricelistItemDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPricelistItemApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesPricelistItemEngineParam) itCatalog.SalesPricelistItemRepository {
			return p.Engine.Repository().(itCatalog.SalesPricelistItemRepository)
		},
		func(p salesPricelistItemEngineParam) itCatalog.SalesPricelistItemDomainService {
			return p.Engine.DomainService().(itCatalog.SalesPricelistItemDomainService)
		},
		func(p salesPricelistItemEngineParam) itCatalog.SalesPricelistItemApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesPricelistItemApplicationService)
		},
	))
	return err
}

type salesComboEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_combo"`
}

// The combo resources carry no rules beyond their schemas, so they take the composable domain
// service unchanged. They still get all four layers: a rule added later has a home, and the layer
// is published by its own type rather than a shared one.
func registerSalesComboEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesComboSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesComboSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesComboRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesComboApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesComboEngineParam) itCatalog.SalesComboRepository {
			return p.Engine.Repository().(itCatalog.SalesComboRepository)
		},
		func(p salesComboEngineParam) itCatalog.SalesComboApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesComboApplicationService)
		},
	))
	return err
}

type salesComboComponentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_combo_component"`
}

func registerSalesComboComponentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesComboComponentSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesComboComponentSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesComboComponentRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesComboComponentApplicationService(base)
				},
			}, param)
		},
	)
	err = stdErr.Join(err, deps.Register(
		func(p salesComboComponentEngineParam) itCatalog.SalesComboComponentRepository {
			return p.Engine.Repository().(itCatalog.SalesComboComponentRepository)
		},
		func(p salesComboComponentEngineParam) itCatalog.SalesComboComponentApplicationService {
			return p.Engine.ApplicationService().(itCatalog.SalesComboComponentApplicationService)
		},
	))
	return err
}
