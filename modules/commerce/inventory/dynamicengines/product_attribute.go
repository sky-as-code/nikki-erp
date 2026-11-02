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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type productAttributeEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute"`
}

// registerProductAttributeEngine declares the product_attribute onion and publishes its typed layers.
func registerProductAttributeEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductAttributeSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductAttributeSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductAttributeRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductAttributeDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
