package core

import (
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	// "github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/shared"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton shared.NikkiModule = &CoreModule{}

type CoreModule struct {
}

// Name implements NikkiModule.
func (*CoreModule) Name() string {
	return "core"
}

// Deps implements NikkiModule.
func (*CoreModule) Deps() []string {
	return nil
}

// Init implements NikkiModule.
func (*CoreModule) Init() error {
	err := config.InitSubModule()
	if err != nil {
		return err
	}
	return nil
}
