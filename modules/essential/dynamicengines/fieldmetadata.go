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
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

type fieldmetadataEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_field_metadata"`
}

func registerFieldMetadataEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.FieldMetadataSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.FieldMetadataSchemaName,
				PermissionScope: tenantScope(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewFieldMetadataRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewFieldMetadataDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
