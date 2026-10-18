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
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

type fieldmetadataEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_field_metadata"`
}

func registerFieldMetadataEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.FieldMetadataSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.FieldMetadataSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewFieldMetadataRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewFieldMetadataDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewFieldMetadataApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p fieldmetadataEngineParam) itFieldMeta.FieldMetadataRepository {
			return p.Engine.Repository().(itFieldMeta.FieldMetadataRepository)
		},
		func(p fieldmetadataEngineParam) itFieldMeta.FieldMetadataDomainService {
			return p.Engine.DomainService().(itFieldMeta.FieldMetadataDomainService)
		},
		func(p fieldmetadataEngineParam) itFieldMeta.FieldMetadataApplicationService {
			return p.Engine.ApplicationService().(itFieldMeta.FieldMetadataApplicationService)
		},
	))
}
