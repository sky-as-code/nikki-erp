package authenticate

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/app"
	repo "github.com/sky-as-code/nikki-erp/modules/authenticate/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &AuthenticateModule{}

type AuthenticateModule struct {
}

// LabelKey implements NikkiModule.
func (*AuthenticateModule) LabelKey() string {
	return "authenticate.moduleLabel"
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

// Version implements NikkiModule.
func (*AuthenticateModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
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
