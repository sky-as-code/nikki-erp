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

type stockMoveDependencyEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move_dependency"`
}

// registerStockMoveDependencyEngine declares the stock_move_dependency onion and publishes its typed layers.
func registerStockMoveDependencyEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockMoveDependencySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockMoveDependencySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockMoveDependencyRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockMoveDependencyDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockMoveDependencyApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockMoveDependencyEngineParam) itStock.StockMoveDependencyRepository {
			return p.Engine.Repository().(itStock.StockMoveDependencyRepository)
		},
		func(p stockMoveDependencyEngineParam) itStock.StockMoveDependencyDomainService {
			return p.Engine.DomainService().(itStock.StockMoveDependencyDomainService)
		},
		func(p stockMoveDependencyEngineParam) itStock.StockMoveDependencyApplicationService {
			return p.Engine.ApplicationService().(itStock.StockMoveDependencyApplicationService)
		},
	))
}
