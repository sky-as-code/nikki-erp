package common

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/config"
	"github.com/sky-as-code/nikki-erp/common/cqrs"
	db "github.com/sky-as-code/nikki-erp/common/database"
	http "github.com/sky-as-code/nikki-erp/common/httpserver"
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules"
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
