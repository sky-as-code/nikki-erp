package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

type uomEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_uom"`
}

// registerUomEngine declares the UoM onion and publishes its typed layers. The repository is
// what the UoM Category onion and the conversion service inject to read UoM rows.
func registerUomEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.UomSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.UomSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewUomRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewUomDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewUomApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p uomEngineParam) itUom.UomRepository {
			return p.Engine.Repository().(itUom.UomRepository)
		},
		func(p uomEngineParam) itUom.UomDomainService {
			return p.Engine.DomainService().(itUom.UomDomainService)
		},
		func(p uomEngineParam) itUom.UomApplicationService {
			return p.Engine.ApplicationService().(itUom.UomApplicationService)
		},
	))
}
