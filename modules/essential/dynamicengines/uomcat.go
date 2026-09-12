package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

type uomCatEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_uomcat"`
}

// registerUomCatEngine declares the UoM Category onion. Its domain service reads UoM rows, so
// the constructor also asks the container for the UoM repository; the container builds the UoM
// onion first, and there is no cycle because the UoM onion needs nothing from the category.
func registerUomCatEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.UomCatSchemaName),
		func(
			param composable.BuildParam,
			uomRepo itUom.UomRepository,
			usageDispatcher *usagecheck.Dispatcher,
		) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.UomCatSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewUomCatRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewUomCatDomainService(base, uomRepo, usageDispatcher)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewUomCatApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p uomCatEngineParam) itUomCat.UomCatRepository {
			return p.Engine.Repository().(itUomCat.UomCatRepository)
		},
		func(p uomCatEngineParam) itUomCat.UomCatDomainService {
			return p.Engine.DomainService().(itUomCat.UomCatDomainService)
		},
		func(p uomCatEngineParam) itUomCat.UomCatApplicationService {
			return p.Engine.ApplicationService().(itUomCat.UomCatApplicationService)
		},
	))
}
