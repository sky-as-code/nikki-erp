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
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type productVariantAttributeValueEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_variant_attribute_value"`
}

// registerProductVariantAttributeValueEngine declares the product_variant_attribute_value onion and publishes its typed layers.
func registerProductVariantAttributeValueEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductVariantAttributeValueSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductVariantAttributeValueSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductVariantAttributeValueRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductVariantAttributeValueDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductVariantAttributeValueApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productVariantAttributeValueEngineParam) itProduct.ProductVariantAttributeValueRepository {
			return p.Engine.Repository().(itProduct.ProductVariantAttributeValueRepository)
		},
		func(p productVariantAttributeValueEngineParam) itProduct.ProductVariantAttributeValueDomainService {
			return p.Engine.DomainService().(itProduct.ProductVariantAttributeValueDomainService)
		},
		func(p productVariantAttributeValueEngineParam) itProduct.ProductVariantAttributeValueApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductVariantAttributeValueApplicationService)
		},
	))
}
