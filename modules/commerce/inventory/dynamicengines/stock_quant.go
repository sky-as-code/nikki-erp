package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type stockQuantEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_quant"`
}

// registerStockQuantEngine declares the stock_quant onion and publishes its typed layers.
func registerStockQuantEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.StockQuantSchemaName),
		func(
			param dynamicresource.BuildParam,
			usageDispatcher *usagecheck.Dispatcher,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockQuantSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewStockQuantRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewStockQuantDomainService(base, usageDispatcher)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewStockQuantApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockQuantEngineParam) itStock.StockQuantRepository {
			return p.Engine.Repository().(itStock.StockQuantRepository)
		},
		func(p stockQuantEngineParam) itStock.StockQuantDomainService {
			return p.Engine.DomainService().(itStock.StockQuantDomainService)
		},
		func(p stockQuantEngineParam) itStock.StockQuantApplicationService {
			return p.Engine.ApplicationService().(itStock.StockQuantApplicationService)
		},
		// The same instance answers what Stock holds at a location, consulted before a location is
		// suspended or archived, and how much of a variant is on hand per warehouse and location.
		func(p stockQuantEngineParam) itStock.LocationUsageReadService {
			return p.Engine.DomainService().(itStock.LocationUsageReadService)
		},
		func(p stockQuantEngineParam) itStock.StockProductSummaryReader {
			return p.Engine.DomainService().(itStock.StockProductSummaryReader)
		},
	))
}
