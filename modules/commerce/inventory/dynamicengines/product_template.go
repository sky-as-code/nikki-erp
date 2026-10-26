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
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type productTemplateEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template"`
}

// registerProductTemplateEngine declares the product_template onion and publishes its typed layers.
func registerProductTemplateEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ProductTemplateSchemaName),
		func(
			param composable.BuildParam,
			usageDispatcher *usagecheck.Dispatcher,
			catalogPublisher itEvent.CatalogChangedEventPublisher,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ProductTemplateSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewProductTemplateRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewProductTemplateDomainService(base, usageDispatcher, composable.NewLazyStorageConsumer())
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewProductTemplateApplicationService(base, catalogPublisher)
				},
				ComputedFunctions: newTemplateComputedFunctions(),
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p productTemplateEngineParam) itProduct.ProductTemplateRepository {
			return p.Engine.Repository().(itProduct.ProductTemplateRepository)
		},
		func(p productTemplateEngineParam) itProduct.ProductTemplateApplicationService {
			return p.Engine.ApplicationService().(itProduct.ProductTemplateApplicationService)
		},
		// ProductTemplateDomainService is a type alias for ProductService (see
		// interfaces/product/product_template_crud.go), so only one provider of the underlying
		// type may be registered. ProductService is the name consumers already inject.
		func(p productTemplateEngineParam) itProduct.ProductService {
			return p.Engine.DomainService().(itProduct.ProductService)
		},
	))
}
