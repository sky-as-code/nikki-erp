package identity

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/identity/app"
	repo "github.com/sky-as-code/nikki-erp/modules/identity/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/identity/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &IdentityModule{}

type IdentityModule struct {
}

// Name implements NikkiModule.
func (*IdentityModule) Name() string {
	return "identity"
}

// Deps implements NikkiModule.
func (*IdentityModule) Deps() []string {
	return []string{}
}

// Init implements NikkiModule.
func (*IdentityModule) Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
