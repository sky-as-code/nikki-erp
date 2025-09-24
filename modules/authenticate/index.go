package authenticate

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app"
	repo "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &AuthenticateModule{}

type AuthenticateModule struct {
}

// Name implements NikkiModule.
func (*AuthenticateModule) Name() string {
	return "authenticate"
}

// Deps implements NikkiModule.
func (*AuthenticateModule) Deps() []string {
	return []string{
		"identity",
	}
}

// Init implements NikkiModule.
func (*AuthenticateModule) Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
