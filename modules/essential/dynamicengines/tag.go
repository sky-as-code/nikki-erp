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
	itTag "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

type tagEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_tag"`
}

func registerTagEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TagSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.TagSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTagRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTagDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTagApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p tagEngineParam) itTag.TagRepository {
			return p.Engine.Repository().(itTag.TagRepository)
		},
		func(p tagEngineParam) itTag.TagDomainService {
			return p.Engine.DomainService().(itTag.TagDomainService)
		},
		func(p tagEngineParam) itTag.TagApplicationService {
			return p.Engine.ApplicationService().(itTag.TagApplicationService)
		},
	))
}
