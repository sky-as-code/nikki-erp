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

type stockTransferEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_transfer"`
}

// registerStockTransferEngine declares the stock_transfer onion and publishes its typed layers.
func registerStockTransferEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockTransferSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockTransferSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockTransferRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockTransferDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockTransferApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockTransferEngineParam) itStock.StockTransferRepository {
			return p.Engine.Repository().(itStock.StockTransferRepository)
		},
		func(p stockTransferEngineParam) itStock.StockTransferDomainService {
			return p.Engine.DomainService().(itStock.StockTransferDomainService)
		},
		func(p stockTransferEngineParam) itStock.StockTransferApplicationService {
			return p.Engine.ApplicationService().(itStock.StockTransferApplicationService)
		},
		// Published for consumers that sequence a movement outside a request, from their own
		// transaction boundaries. The narrowed interface is published, never the struct: handing
		// over the embedded CRUD would make the lifecycle rules optional.
		func(p stockTransferEngineParam) itStock.StockTransferMovementService {
			return p.Engine.DomainService().(itStock.StockTransferMovementService)
		},
	))
}
