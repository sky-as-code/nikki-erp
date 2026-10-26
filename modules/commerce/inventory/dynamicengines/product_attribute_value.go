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
