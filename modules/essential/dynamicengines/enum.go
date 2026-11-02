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
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

type enumEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_enum"`
}

func registerEnumEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.EnumSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.EnumSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewEnumRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewEnumDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
