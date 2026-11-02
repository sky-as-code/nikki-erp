package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/settings/infra/repository"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

type settingsSchemaEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_settings_schema"`
}

func registerSettingsSchemaEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SettingsSchemaSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SettingsSchemaSchemaName,
				CrudActions: ReadOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSettingsSchemaRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSettingsSchemaDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
