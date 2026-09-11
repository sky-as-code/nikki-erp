package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/inventory/infra/repository"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type stockProductConfigEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_product_config"`
}

// registerStockProductConfigEngine declares the stock_product_config onion and publishes its typed layers.
func registerStockProductConfigEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockProductConfigSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockProductConfigSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockProductConfigRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockProductConfigDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockProductConfigApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockProductConfigEngineParam) itStock.StockProductConfigRepository {
			return p.Engine.Repository().(itStock.StockProductConfigRepository)
		},
		func(p stockProductConfigEngineParam) itStock.StockProductConfigDomainService {
			return p.Engine.DomainService().(itStock.StockProductConfigDomainService)
		},
		func(p stockProductConfigEngineParam) itStock.StockProductConfigApplicationService {
			return p.Engine.ApplicationService().(itStock.StockProductConfigApplicationService)
		},
	))
}
