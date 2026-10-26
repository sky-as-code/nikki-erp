package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type productAttributeValueEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute_value"`
}

// registerProductAttributeValueEngine declares the product_attribute_value onion and publishes its typed layers.
func registerProductAttributeValueEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductAttributeValueSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductAttributeValueSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductAttributeValueRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductAttributeValueDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductAttributeValueApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productAttributeValueEngineParam) itProduct.ProductAttributeValueRepository {
			return p.Engine.Repository().(itProduct.ProductAttributeValueRepository)
		},
		func(p productAttributeValueEngineParam) itProduct.ProductAttributeValueDomainService {
			return p.Engine.DomainService().(itProduct.ProductAttributeValueDomainService)
		},
		func(p productAttributeValueEngineParam) itProduct.ProductAttributeValueApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductAttributeValueApplicationService)
		},
	))
}
