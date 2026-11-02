package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/services"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
	itQuota "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsquota"
)

type settingsQuotaEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_settings_quota"`
}

// registerSettingsQuotaEngine needs no org scoping: a quota row has no org_id, and its owner may
// be the tenant itself.
func registerSettingsQuotaEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SettingsQuotaSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SettingsQuotaSchemaName,
				CrudActions: ReadOnlyCrudActions(),
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSettingsQuotaDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSettingsQuotaApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p settingsQuotaEngineParam) itQuota.SettingsQuotaDomainService {
			return p.Engine.DomainService().(itQuota.SettingsQuotaDomainService)
		},
		func(p settingsQuotaEngineParam) itQuota.SettingsQuotaApplicationService {
			return p.Engine.ApplicationService().(itQuota.SettingsQuotaApplicationService)
		},
		func(p settingsQuotaEngineParam) itSettings.QuotaAppService {
			return p.Engine.ApplicationService().(itQuota.SettingsQuotaApplicationService)
		},
	))
}
