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
	itModelMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

type modelmetadataEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_model_metadata"`
}

func registerModelMetadataEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ModelMetadataSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ModelMetadataSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewModelMetadataRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewModelMetadataDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewModelMetadataApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p modelmetadataEngineParam) itModelMeta.ModelMetadataRepository {
			return p.Engine.Repository().(itModelMeta.ModelMetadataRepository)
		},
		func(p modelmetadataEngineParam) itModelMeta.ModelMetadataDomainService {
			return p.Engine.DomainService().(itModelMeta.ModelMetadataDomainService)
		},
		func(p modelmetadataEngineParam) itModelMeta.ModelMetadataApplicationService {
			return p.Engine.ApplicationService().(itModelMeta.ModelMetadataApplicationService)
		},
	))
}
