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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type stockMoveEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move"`
}

// registerStockMoveEngine declares the stock_move onion and publishes its typed layers.
func registerStockMoveEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.StockMoveSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockMoveSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewStockMoveRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewStockMoveDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
