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
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

type enumEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_enum"`
}

func registerEnumEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.EnumSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.EnumSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewEnumRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewEnumDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewEnumApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p enumEngineParam) itEnum.EnumRepository {
			return p.Engine.Repository().(itEnum.EnumRepository)
		},
		func(p enumEngineParam) itEnum.EnumDomainService {
			return p.Engine.DomainService().(itEnum.EnumDomainService)
		},
		func(p enumEngineParam) itEnum.EnumApplicationService {
			return p.Engine.ApplicationService().(itEnum.EnumApplicationService)
		},
	))
}
