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

type productTemplateAttributeValueEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute_value"`
}

// registerProductTemplateAttributeValueEngine declares the product_template_attribute_value onion and publishes its typed layers.
func registerProductTemplateAttributeValueEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductTemplateAttributeValueSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateAttributeValueSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductTemplateAttributeValueRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductTemplateAttributeValueDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
