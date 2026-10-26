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

type productAttributeEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute"`
}

// registerProductAttributeEngine declares the product_attribute onion and publishes its typed layers.
func registerProductAttributeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductAttributeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductAttributeSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductAttributeRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductAttributeDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductAttributeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productAttributeEngineParam) itProduct.ProductAttributeRepository {
			return p.Engine.Repository().(itProduct.ProductAttributeRepository)
		},
		func(p productAttributeEngineParam) itProduct.ProductAttributeDomainService {
			return p.Engine.DomainService().(itProduct.ProductAttributeDomainService)
		},
		func(p productAttributeEngineParam) itProduct.ProductAttributeApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductAttributeApplicationService)
		},
	))
}
