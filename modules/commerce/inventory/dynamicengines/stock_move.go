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

type stockMoveEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move"`
}

// registerStockMoveEngine declares the stock_move onion and publishes its typed layers.
func registerStockMoveEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockMoveSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockMoveSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockMoveRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockMoveDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockMoveApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockMoveEngineParam) itStock.StockMoveRepository {
			return p.Engine.Repository().(itStock.StockMoveRepository)
		},
		func(p stockMoveEngineParam) itStock.StockMoveDomainService {
			return p.Engine.DomainService().(itStock.StockMoveDomainService)
		},
		func(p stockMoveEngineParam) itStock.StockMoveApplicationService {
			return p.Engine.ApplicationService().(itStock.StockMoveApplicationService)
		},
	))
}
