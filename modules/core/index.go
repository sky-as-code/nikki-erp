package core

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	http "github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &CoreModule{}

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
	err := errors.Join(
		deps.Invoke(config.InitSubModule),
		deps.Invoke(cqrs.InitSubModule),
		deps.Register(db.InitSubModule),
		deps.Register(http.InitSubModule),
	)

	if err != nil {
		return err
	}
	err = errors.Join(
		deps.Invoke(db.InitSubModule),
	)
	return err
}
