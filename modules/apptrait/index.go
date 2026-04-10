package apptrait

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/config"
	"github.com/sky-as-code/nikki-erp/modules"
	httpserverExt "github.com/sky-as-code/nikki-erp/modules/core/httpserver/external"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &AppTraitModule{}

// This is the module to configure some settings specific to each application build.
// It is not managed by the module system.
type AppTraitModule struct {
}

// LabelKey implements NikkiModule.
func (*AppTraitModule) LabelKey() string {
	return "apptrait.moduleLabel"
}

// Name implements NikkiModule.
func (*AppTraitModule) Name() string {
	return "apptrait"
}

// Deps implements NikkiModule.
func (*AppTraitModule) Deps() []string {
	return []string{}
}

// Version implements NikkiModule.
func (*AppTraitModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*AppTraitModule) Init() error {
	err := stdErr.Join(
		deps.Register(httpserverExt.NewPermissionExtServiceImpl),
		deps.Register(requestguard.NewStaticRequestGuardServiceImpl),
		deps.Register(config.GetDefaultConfigYaml),
	)

	return err
}

// RegisterModels implements DynamicModule.
func (*AppTraitModule) RegisterModels() error {
	return nil
}
