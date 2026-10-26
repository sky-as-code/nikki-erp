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

type brandEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_brand"`
}

// registerBrandEngine declares the brand onion and publishes its typed layers.
func registerBrandEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.BrandSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.BrandSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewBrandRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewBrandDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewBrandApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p brandEngineParam) itProduct.BrandRepository {
			return p.Engine.Repository().(itProduct.BrandRepository)
		},
		func(p brandEngineParam) itProduct.BrandDomainService {
			return p.Engine.DomainService().(itProduct.BrandDomainService)
		},
		func(p brandEngineParam) itProduct.BrandApplicationService {
			return p.Engine.ApplicationService().(itProduct.BrandApplicationService)
		},
	))
}
