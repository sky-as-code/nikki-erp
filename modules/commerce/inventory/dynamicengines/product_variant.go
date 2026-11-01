package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type productVariantEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_variant"`
}

// registerProductVariantEngine declares the product_variant onion and publishes its typed layers.
func registerProductVariantEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductVariantSchemaName),
		func(
			param dynamicresource.BuildParam,
			productSvc itProduct.ProductService,
			usageDispatcher *usagecheck.Dispatcher,
			catalogPublisher itEvent.CatalogChangedEventPublisher,
			units *services.OrgUnitResolver,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductVariantSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductVariantRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductVariantDomainService(base, usageDispatcher)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewProductVariantApplicationService(base, catalogPublisher, productSvc)
				},
				// Resolved lazily, per call: the attribute-value onion is registered by this same
				// pass, so looking it up while building this one would depend on registration order.
				ComputedFunctions: mergeComputedFunctions(
					newVariantComputedFunctions(dynamicresource.LookupSourceOnion), newVariantUnitFunctions(units)),
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productVariantEngineParam) itProduct.ProductVariantRepository {
			return p.Engine.Repository().(itProduct.ProductVariantRepository)
		},
		func(p productVariantEngineParam) itProduct.ProductVariantDomainService {
			return p.Engine.DomainService().(itProduct.ProductVariantDomainService)
		},
		func(p productVariantEngineParam) itProduct.ProductVariantApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductVariantApplicationService)
		},
		// One instance serves all these ports, so a consumer gets the batched template_* fill
		// whichever it injects. The pricing-basis port stays separate because it grants strictly
		// less: a price calculator gets the pricing inputs without a general product reader.
		func(p productVariantEngineParam) itProduct.ProductTemplateReadService {
			return p.Engine.DomainService().(itProduct.ProductTemplateReadService)
		},
		func(p productVariantEngineParam) itProduct.ProductCategoryReadService {
			return p.Engine.DomainService().(itProduct.ProductCategoryReadService)
		},
		func(p productVariantEngineParam) itProduct.ProductPricingBasisService {
			return p.Engine.DomainService().(itProduct.ProductPricingBasisService)
		},
		// The org-scoped read, for a caller that reaches this port with no user permission of its
		// own and so must supply the org itself rather than have it asserted from the caller's
		// membership.
		func(domain itProduct.ProductVariantDomainService) itProduct.ProductVariantService {
			return app.NewProductVariantReadService(domain)
		},
	))
}
