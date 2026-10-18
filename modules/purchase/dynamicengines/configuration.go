package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/purchase/infra/repository"
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/configuration"
)

// The per-org purchase configuration.

type configurationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_configuration"`
}

func registerConfigurationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ConfigurationSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ConfigurationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewConfigurationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewConfigurationCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewConfigurationApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p configurationEngineParam) itConfiguration.ConfigurationCrudRepository {
			return p.Engine.Repository().(itConfiguration.ConfigurationCrudRepository)
		},
		func(p configurationEngineParam) itConfiguration.ConfigurationCrudDomainService {
			return p.Engine.DomainService().(itConfiguration.ConfigurationCrudDomainService)
		},
		func(p configurationEngineParam) itConfiguration.ConfigurationCrudApplicationService {
			return p.Engine.ApplicationService().(itConfiguration.ConfigurationCrudApplicationService)
		},
	))
}
