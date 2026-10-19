package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/settings/infra/repository"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

type settingsSchemaEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_settings_schema"`
}

func registerSettingsSchemaEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SettingsSchemaSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SettingsSchemaSchemaName,
				CrudActions: ReadOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSettingsSchemaRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSettingsSchemaDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSettingsSchemaApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p settingsSchemaEngineParam) itSchema.SettingsSchemaRepository {
			return p.Engine.Repository().(itSchema.SettingsSchemaRepository)
		},
		func(p settingsSchemaEngineParam) itSchema.SettingsSchemaDomainService {
			return p.Engine.DomainService().(itSchema.SettingsSchemaDomainService)
		},
		func(p settingsSchemaEngineParam) itSchema.SettingsSchemaApplicationService {
			return p.Engine.ApplicationService().(itSchema.SettingsSchemaApplicationService)
		},
	))
}
