package contacts

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/contacts/app"
	"github.com/sky-as-code/nikki-erp/modules/contacts/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/contacts/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.NikkiModule = &ContactsModule{}

type ContactsModule struct {
}

// Name implements NikkiModule.
func (*ContactsModule) Name() string {
	return "contacts"
}

// Deps implements NikkiModule.
func (*ContactsModule) Deps() []string {
	return []string{}
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
