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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type stockScrapEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_scrap"`
}

// registerStockScrapEngine declares the stock_scrap onion and publishes its typed layers.
func registerStockScrapEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockScrapSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockScrapSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockScrapRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockScrapDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockScrapApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockScrapEngineParam) itStock.StockScrapRepository {
			return p.Engine.Repository().(itStock.StockScrapRepository)
		},
		func(p stockScrapEngineParam) itStock.StockScrapDomainService {
			return p.Engine.DomainService().(itStock.StockScrapDomainService)
		},
		func(p stockScrapEngineParam) itStock.StockScrapApplicationService {
			return p.Engine.ApplicationService().(itStock.StockScrapApplicationService)
		},
	))
}
