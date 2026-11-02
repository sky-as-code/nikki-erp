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

type productTemplateCombinationExclusionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_combination_exclusion"`
}

// registerProductTemplateCombinationExclusionEngine declares the exclusion onion and publishes its
// typed layers. Its built-in writes stay on the domain service for the template configuration;
// the REST table exposes reads only.
func registerProductTemplateCombinationExclusionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductTemplateCombinationExclusionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateCombinationExclusionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductTemplateCombinationExclusionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductTemplateCombinationExclusionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewProductTemplateCombinationExclusionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productTemplateCombinationExclusionEngineParam) itProduct.ProductTemplateCombinationExclusionRepository {
			return p.Engine.Repository().(itProduct.ProductTemplateCombinationExclusionRepository)
		},
		func(p productTemplateCombinationExclusionEngineParam) itProduct.ProductTemplateCombinationExclusionDomainService {
			return p.Engine.DomainService().(itProduct.ProductTemplateCombinationExclusionDomainService)
		},
		func(p productTemplateCombinationExclusionEngineParam) itProduct.ProductTemplateCombinationExclusionApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductTemplateCombinationExclusionApplicationService)
		},
	))
}
