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

type productCategoryEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_category"`
}

// registerProductCategoryEngine declares the product_category onion and publishes its typed layers.
func registerProductCategoryEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductCategorySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductCategorySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductCategoryRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductCategoryDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductCategoryApplicationService(base)
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
