package authorize

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	app "github.com/sky-as-code/nikki-erp/modules/authorize/app"
	repo "github.com/sky-as-code/nikki-erp/modules/authorize/infra/repository"
	transport "github.com/sky-as-code/nikki-erp/modules/authorize/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &AuthorizeModule{}

type AuthorizeModule struct {
}

// LabelKey implements NikkiModule.
func (*AuthorizeModule) LabelKey() string {
	return "authorize.moduleLabel"
}

// Name implements NikkiModule.
func (*AuthorizeModule) Name() string {
	return "authorize"
}

// Deps implements NikkiModule.
func (*AuthorizeModule) Deps() []string {
	return []string{
		"identity",
	}
}

// Version implements NikkiModule.
func (*AuthorizeModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*AuthorizeModule) Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
