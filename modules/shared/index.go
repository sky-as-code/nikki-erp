package shared

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/shared/config"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &SharedModule{}

type SharedModule struct {
}

// Name implements NikkiModule.
func (*SharedModule) Name() string {
	return "shared"
}

// Deps implements NikkiModule.
func (*SharedModule) Deps() []string {
	return nil
}

// Init implements NikkiModule.
func (*SharedModule) Init() error {
	err := config.InitSubModule()
	if err != nil {
		return err
	}
	return nil
}
