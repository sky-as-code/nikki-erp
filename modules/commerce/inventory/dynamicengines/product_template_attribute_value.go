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

type productTemplateAttributeValueEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute_value"`
}

// registerProductTemplateAttributeValueEngine declares the product_template_attribute_value onion and publishes its typed layers.
func registerProductTemplateAttributeValueEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductTemplateAttributeValueSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateAttributeValueSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductTemplateAttributeValueRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductTemplateAttributeValueDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewProductTemplateAttributeValueApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productTemplateAttributeValueEngineParam) itProduct.ProductTemplateAttributeValueRepository {
			return p.Engine.Repository().(itProduct.ProductTemplateAttributeValueRepository)
		},
		func(p productTemplateAttributeValueEngineParam) itProduct.ProductTemplateAttributeValueDomainService {
			return p.Engine.DomainService().(itProduct.ProductTemplateAttributeValueDomainService)
		},
		func(p productTemplateAttributeValueEngineParam) itProduct.ProductTemplateAttributeValueApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductTemplateAttributeValueApplicationService)
		},
	))
}
