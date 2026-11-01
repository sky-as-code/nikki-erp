package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itTag "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

type tagEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_tag"`
}

func registerTagEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TagSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.TagSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTagRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTagDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
