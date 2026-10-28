package dynamicresource

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ModuleName is the name feature modules must list in their Deps() to use composable resource onions.
const ModuleName = "dynamicresource"

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &DynamicResourceModule{}

type DynamicResourceModule struct {
}

// Name implements InCodeModule.
func (*DynamicResourceModule) Name() string {
	return ModuleName
}

// ModelPrefixes implements DynamicModule.
func (*DynamicResourceModule) ModelPrefixes() []string {
	return []string{ModuleName}
}

func (*DynamicResourceModule) HasUi() bool {
	return false
}

func (*DynamicResourceModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryFoundation
}

// LabelKey implements InCodeModule.
func (*DynamicResourceModule) LabelKey() string {
	return "dynamicresource.moduleLabel"
}

// Deps implements InCodeModule.
func (*DynamicResourceModule) Deps() []string {
	return []string{
		"core",
	}
}

// IsInternal implements InCodeModule.
func (*DynamicResourceModule) IsInternal() bool {
	return true
}

// Version implements InCodeModule.
func (*DynamicResourceModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements InCodeModule. The engine registry answers core/resourceaccess itself: every
// engine resource is reachable from it, so there is no feature module to bind the port to.
func (*DynamicResourceModule) Init() error {
	return stdErr.Join(
		orglifecycle.RegisterNoop(ModuleName),
		deps.Register(composable.NewResourceAccess),
	)
}

// OnAppStarted implements InCodeModuleAppStarted.
//
// It matches every "function"-kind computed field against the functions its onion registered.
// The check has to run here rather than at onion construction: a feature module builds its
// onions and registers their functions inside its own Init(), which is long after this module's.
// By the time any OnAppStarted runs, every module's Init() has completed — so a missing function
// fails the boot with a precise message instead of surfacing as a broken read later.
func (*DynamicResourceModule) OnAppStarted() error {
	return composable.AssertComputedFunctionsDefined()
}
