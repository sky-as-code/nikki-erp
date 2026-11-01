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

type stockMoveLineEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move_line"`
}

// registerStockMoveLineEngine declares the stock_move_line onion and publishes its typed layers.
func registerStockMoveLineEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.StockMoveLineSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockMoveLineSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewStockMoveLineRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewStockMoveLineDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewStockMoveLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockMoveLineEngineParam) itStock.StockMoveLineRepository {
			return p.Engine.Repository().(itStock.StockMoveLineRepository)
		},
		func(p stockMoveLineEngineParam) itStock.StockMoveLineDomainService {
			return p.Engine.DomainService().(itStock.StockMoveLineDomainService)
		},
		func(p stockMoveLineEngineParam) itStock.StockMoveLineApplicationService {
			return p.Engine.ApplicationService().(itStock.StockMoveLineApplicationService)
		},
	))
}
