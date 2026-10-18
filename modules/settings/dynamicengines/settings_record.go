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
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

type settingsRecordEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_settings_record"`
}

func registerSettingsRecordEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SettingsRecordSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SettingsRecordSchemaName,
				CrudActions: ReadOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSettingsRecordRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSettingsRecordDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSettingsRecordApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p settingsRecordEngineParam) itRecord.SettingsRecordRepository {
			return p.Engine.Repository().(itRecord.SettingsRecordRepository)
		},
		func(p settingsRecordEngineParam) itRecord.SettingsRecordDomainService {
			return p.Engine.DomainService().(itRecord.SettingsRecordDomainService)
		},
		func(p settingsRecordEngineParam) itRecord.SettingsRecordApplicationService {
			return p.Engine.ApplicationService().(itRecord.SettingsRecordApplicationService)
		},
	))
}
