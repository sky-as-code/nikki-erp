package contacts

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/contacts/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &ContactsModule{}

type ContactsModule struct {
}

// LabelKey implements NikkiModule.
func (*ContactsModule) LabelKey() string {
	return "contacts.moduleLabel"
}

// Name implements NikkiModule.
func (*ContactsModule) Name() string {
	return "contacts"
}

// Deps implements NikkiModule.
func (*ContactsModule) Deps() []string {
	return []string{}
}

// Version implements NikkiModule.
func (*ContactsModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*ContactsModule) Init() error {
	err := errors.Join(
		repository.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
