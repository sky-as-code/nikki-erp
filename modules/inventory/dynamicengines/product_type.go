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

type productTypeEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_type"`
}

// registerProductTypeEngine declares the product_type onion and publishes its typed layers.
func registerProductTypeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductTypeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTypeSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductTypeRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductTypeDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductTypeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productTypeEngineParam) itProduct.ProductTypeRepository {
			return p.Engine.Repository().(itProduct.ProductTypeRepository)
		},
		func(p productTypeEngineParam) itProduct.ProductTypeDomainService {
			return p.Engine.DomainService().(itProduct.ProductTypeDomainService)
		},
		func(p productTypeEngineParam) itProduct.ProductTypeApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductTypeApplicationService)
		},
	))
}
