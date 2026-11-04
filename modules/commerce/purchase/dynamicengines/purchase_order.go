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

// The purchase order (an RFQ and a PO are one record at different statuses), with its twelve
// lifecycle actions. The ports are what the order's create and reprice ask.

type purchaseOrderEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_purchase_order"`
}

func registerPurchaseOrderEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.PurchaseOrderSchemaName),
		func(
			param dynamicresource.BuildParam,
			vendors itExt.VendorExtService,
			currencies itExt.CurrencyExtService,
			products itExt.ProductExtService,
			uoms itExt.UomExtService,
			numbering *services.CodeNumbering,
		) dynamicresource.DynamicResourceEngineOnion {
			references := services.NewOrderReferenceValidator(vendors, currencies)
			// Totals round to the order's own currency rather than a fixed two places.
			services.SetOrderScaleResolver(references.ScaleFor)
			validator := services.NewProductLineValidator(products, uoms)
			pricer := services.NewLinePricer(uoms)
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.PurchaseOrderSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewPurchaseOrderRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewPurchaseOrderDomainService(base, references, validator, pricer, numbering)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewPurchaseOrderApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p purchaseOrderEngineParam) itOrder.PurchaseOrderCrudRepository {
			return p.Engine.Repository().(itOrder.PurchaseOrderCrudRepository)
		},
		func(p purchaseOrderEngineParam) itOrder.PurchaseOrderCrudDomainService {
			return p.Engine.DomainService().(itOrder.PurchaseOrderCrudDomainService)
		},
		func(p purchaseOrderEngineParam) itOrder.PurchaseOrderCrudApplicationService {
			return p.Engine.ApplicationService().(itOrder.PurchaseOrderCrudApplicationService)
		},
	))
}
