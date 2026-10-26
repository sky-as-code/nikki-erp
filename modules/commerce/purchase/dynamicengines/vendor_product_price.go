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
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendorprice"
)

// The vendor price list. Its domain service validates a price as it will be on create and update,
// and revalidates on unarchive; the three ports are what the validator asks.

type vendorProductPriceEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_vendor_product_price"`
}

func registerVendorProductPriceEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.VendorProductPriceSchemaName),
		func(
			param composable.BuildParam,
			vendors itExt.VendorExtService,
			products itExt.ProductExtService,
			uoms itExt.UomExtService,
		) composable.DynamicResourceEngineOnion {
			validator := services.NewVendorPriceValidator(vendors, products, uoms)
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.VendorProductPriceSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewVendorProductPriceRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewVendorProductPriceDomainService(base, validator)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewVendorProductPriceApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p vendorProductPriceEngineParam) itVendorPrice.VendorProductPriceCrudRepository {
			return p.Engine.Repository().(itVendorPrice.VendorProductPriceCrudRepository)
		},
		func(p vendorProductPriceEngineParam) itVendorPrice.VendorProductPriceCrudDomainService {
			return p.Engine.DomainService().(itVendorPrice.VendorProductPriceCrudDomainService)
		},
		func(p vendorProductPriceEngineParam) itVendorPrice.VendorProductPriceCrudApplicationService {
			return p.Engine.ApplicationService().(itVendorPrice.VendorProductPriceCrudApplicationService)
		},
	))
}
