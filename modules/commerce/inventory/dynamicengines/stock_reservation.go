package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type stockReservationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_reservation"`
}

// registerStockReservationEngine declares the reservation onion. A reservation is transactional
// data written only by the reservation operations, so the built-in create, update and delete are
// withheld here; REST withholds the same routes in transport/restful.
func registerStockReservationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StockReservationSchemaName),
		func(param composable.BuildParam, uom itExt.UomConversionExtService) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.StockReservationSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStockReservationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStockReservationDomainService(base, uom)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStockReservationApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p stockReservationEngineParam) itStock.StockReservationRepository {
			return p.Engine.Repository().(itStock.StockReservationRepository)
		},
		func(p stockReservationEngineParam) itStock.StockReservationDomainService {
			return p.Engine.DomainService().(itStock.StockReservationDomainService)
		},
		func(p stockReservationEngineParam) itStock.StockReservationApplicationService {
			return p.Engine.ApplicationService().(itStock.StockReservationApplicationService)
		},
		// The reservation operations, as the port other modules bind to.
		func(p stockReservationEngineParam) itStock.WarehouseReservationService {
			return p.Engine.DomainService().(itStock.WarehouseReservationService)
		},
	))
}
