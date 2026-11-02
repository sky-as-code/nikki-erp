package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/purchase/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The purchase order line. Its domain service validates the product and unit, prices the line and
// recomputes the order header in the same transaction.

type purchaseOrderLineEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_purchase_order_line"`
}

func registerPurchaseOrderLineEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.PurchaseOrderLineSchemaName),
		func(
			param dynamicresource.BuildParam,
			products itExt.ProductExtService,
			uoms itExt.UomExtService,
		) dynamicresource.DynamicResourceEngineOnion {
			validator := services.NewProductLineValidator(products, uoms)
			pricer := services.NewLinePricer(uoms)
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.PurchaseOrderLineSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewPurchaseOrderLineRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewPurchaseOrderLineDomainService(base, validator, pricer)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
