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

type productTemplateAttributeEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute"`
}

// registerProductTemplateAttributeEngine declares the product_template_attribute onion and publishes its typed layers.
func registerProductTemplateAttributeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductTemplateAttributeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateAttributeSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductTemplateAttributeRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductTemplateAttributeDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
