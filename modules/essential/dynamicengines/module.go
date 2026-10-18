package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itModule "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

type moduleEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_module_metadata"`
}

// moduleBuildParam is composable.BuildParam plus what the sync and the label rule need.
type moduleBuildParam struct {
	composable.BuildParam

	CqrsBus cqrs.CqrsBus
	Logger  logging.LoggerService
}

// ModuleReadActions is the built-in action set the module catalogue serves over REST. The onion
// keeps every action, because the boot-time sync creates and updates through the domain service.
func ModuleReadActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionGetById, composable.CrudActionSearch,
		composable.CrudActionExists, composable.CrudActionGetSchema,
	}
}

func registerModuleEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ModuleMetadataSchemaName),
		func(param moduleBuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ModuleMetadataSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewModuleRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewModuleDomainService(base, param.CqrsBus, param.Logger)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewModuleApplicationService(base)
				},
			}, param.BuildParam)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p moduleEngineParam) itModule.ModuleRepository {
			return p.Engine.Repository().(itModule.ModuleRepository)
		},
		func(p moduleEngineParam) itModule.ModuleDomainService {
			return p.Engine.DomainService().(itModule.ModuleDomainService)
		},
		func(p moduleEngineParam) itModule.ModuleApplicationService {
			return p.Engine.ApplicationService().(itModule.ModuleApplicationService)
		},
		func(p moduleEngineParam) itModule.ModuleAppService {
			return p.Engine.ApplicationService().(itModule.ModuleAppService)
		},
	))
}
