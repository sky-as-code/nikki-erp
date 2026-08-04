package dynamicresource

import (
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
)

// ModuleName is the name feature modules must list in their Deps() to use resource engines.
const ModuleName = "dynamicresource"

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &DynamicResourceModule{}

type DynamicResourceModule struct {
}

// Name implements InCodeModule.
func (*DynamicResourceModule) Name() string {
	return ModuleName
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

// Init implements InCodeModule.
// It hands the core services that every engine needs to the registry, so that feature
// modules can create their engines during their own Init(). This module is initialized
// before them because they declare it in their Deps().
func (*DynamicResourceModule) Init() error {
	return initRegistryDeps()
}
