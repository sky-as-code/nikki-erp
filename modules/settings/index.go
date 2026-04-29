package settings

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/settings/app"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	services "github.com/sky-as-code/nikki-erp/modules/settings/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/settings/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/settings/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
var ModuleSingleton modules.DynamicModule = &SettingsModule{}

type SettingsModule struct{}

func (*SettingsModule) LabelKey() string {
	return "settings.moduleLabel"
}

func (*SettingsModule) Name() string {
	return "settings"
}

func (*SettingsModule) Deps() []string {
	return []string{}
}

func (*SettingsModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*SettingsModule) Init() error {
	return stdErr.Join(
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

func (*SettingsModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(domain.UserPreferenceSchemaBuilder()),
	)
}
