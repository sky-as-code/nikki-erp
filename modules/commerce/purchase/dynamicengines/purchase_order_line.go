package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/purchase/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

// The purchase order line. Its domain service validates the product and unit, prices the line and
// recomputes the order header in the same transaction.

type purchaseOrderLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_order_line"`
}

func registerPurchaseOrderLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.PurchaseOrderLineSchemaName),
		func(
			param composable.BuildParam,
			products itExt.ProductExtService,
			uoms itExt.UomExtService,
		) composable.DynamicResourceEngineOnion {
			validator := services.NewProductLineValidator(products, uoms)
			pricer := services.NewLinePricer(uoms)
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.PurchaseOrderLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewPurchaseOrderLineRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewPurchaseOrderLineDomainService(base, validator, pricer)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewPurchaseOrderLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p purchaseOrderLineEngineParam) itOrder.PurchaseOrderLineCrudRepository {
			return p.Engine.Repository().(itOrder.PurchaseOrderLineCrudRepository)
		},
		func(p purchaseOrderLineEngineParam) itOrder.PurchaseOrderLineCrudDomainService {
			return p.Engine.DomainService().(itOrder.PurchaseOrderLineCrudDomainService)
		},
		func(p purchaseOrderLineEngineParam) itOrder.PurchaseOrderLineCrudApplicationService {
			return p.Engine.ApplicationService().(itOrder.PurchaseOrderLineCrudApplicationService)
		},
	))
}
