package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type productCategoryEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_category"`
}

// registerProductCategoryEngine declares the product_category onion and publishes its typed layers.
func registerProductCategoryEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ProductCategorySchemaName),
		func(
			param dynamicresource.BuildParam, catalogPublisher itEvent.CatalogChangedEventPublisher,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductCategorySchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewProductCategoryRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewProductCategoryDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewProductCategoryApplicationService(base, catalogPublisher)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productCategoryEngineParam) itProduct.ProductCategoryRepository {
			return p.Engine.Repository().(itProduct.ProductCategoryRepository)
		},
		func(p productCategoryEngineParam) itProduct.ProductCategoryDomainService {
			return p.Engine.DomainService().(itProduct.ProductCategoryDomainService)
		},
		func(p productCategoryEngineParam) itProduct.ProductCategoryApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductCategoryApplicationService)
		},
	))
}
