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

type stockOperationTypeEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_operation_type"`
}

// registerStockOperationTypeEngine declares the stock_operation_type onion and publishes its typed layers.
func registerStockOperationTypeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockOperationTypeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StockOperationTypeSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockOperationTypeRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockOperationTypeDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockOperationTypeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockOperationTypeEngineParam) itStock.StockOperationTypeRepository {
			return p.Engine.Repository().(itStock.StockOperationTypeRepository)
		},
		func(p stockOperationTypeEngineParam) itStock.StockOperationTypeDomainService {
			return p.Engine.DomainService().(itStock.StockOperationTypeDomainService)
		},
		func(p stockOperationTypeEngineParam) itStock.StockOperationTypeApplicationService {
			return p.Engine.ApplicationService().(itStock.StockOperationTypeApplicationService)
		},
	))
}
