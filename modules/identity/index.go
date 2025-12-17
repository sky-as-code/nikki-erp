package identity

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/module"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules/identity/app"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	repo "github.com/sky-as-code/nikki-erp/modules/identity/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/identity/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton module.InCodeModule = &IdentityModule{}

type IdentityModule struct {
}

// LabelKey implements InCodeModule.
func (*IdentityModule) LabelKey() string {
	return "identity.moduleLabel"
}

// Name implements InCodeModule.
func (*IdentityModule) Name() string {
	return "identity"
}

// Deps implements InCodeModule.
func (*IdentityModule) Deps() []string {
	return []string{}
}

// Version implements InCodeModule.
func (*IdentityModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements InCodeModule.

func (this *IdentityModule) Init(opts module.ModuleInitOptions) error {
	opts.RegisterSchema(domain.UserSchemaBuilder(), this.Name())

	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
