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

type productTemplateAttributeEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute"`
}

// registerProductTemplateAttributeEngine declares the product_template_attribute onion and publishes its typed layers.
func registerProductTemplateAttributeEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductTemplateAttributeSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateAttributeSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductTemplateAttributeRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductTemplateAttributeDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewProductTemplateAttributeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productTemplateAttributeEngineParam) itProduct.ProductTemplateAttributeRepository {
			return p.Engine.Repository().(itProduct.ProductTemplateAttributeRepository)
		},
		func(p productTemplateAttributeEngineParam) itProduct.ProductTemplateAttributeDomainService {
			return p.Engine.DomainService().(itProduct.ProductTemplateAttributeDomainService)
		},
		func(p productTemplateAttributeEngineParam) itProduct.ProductTemplateAttributeApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductTemplateAttributeApplicationService)
		},
	))
}
