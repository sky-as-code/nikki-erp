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
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/vendorprice"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The vendor price list. Its domain service validates a price as it will be on create and update,
// and revalidates on unarchive; the three ports are what the validator asks.

type vendorProductPriceEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_purchase_vendor_product_price"`
}

func registerVendorProductPriceEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.VendorProductPriceSchemaName),
		func(
			param dynamicresource.BuildParam,
			vendors itExt.VendorExtService,
			products itExt.ProductExtService,
			uoms itExt.UomExtService,
		) dynamicresource.DynamicResourceEngineOnion {
			validator := services.NewVendorPriceValidator(vendors, products, uoms)
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.VendorProductPriceSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewVendorProductPriceRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewVendorProductPriceDomainService(base, validator)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
