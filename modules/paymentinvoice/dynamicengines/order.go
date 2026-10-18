package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/repository"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

// The payment order, with the create_payment, refund and remove_pos_orders actions, which
// delegate to the order domain service.

type orderEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_order"`
}

func registerOrderEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.OrderSchemaName),
		func(
			param composable.BuildParam,
			orders *services.OrderDomainService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.OrderSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewOrderRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewOrderCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewOrderApplicationService(base, orders)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p orderEngineParam) itOrder.OrderCrudRepository {
			return p.Engine.Repository().(itOrder.OrderCrudRepository)
		},
		func(p orderEngineParam) itOrder.OrderCrudDomainService {
			return p.Engine.DomainService().(itOrder.OrderCrudDomainService)
		},
		func(p orderEngineParam) itOrder.OrderCrudApplicationService {
			return p.Engine.ApplicationService().(itOrder.OrderCrudApplicationService)
		},
	))
}
