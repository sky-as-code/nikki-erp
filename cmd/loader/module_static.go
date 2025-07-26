//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/authenticate"
	"github.com/sky-as-code/nikki-erp/modules/authorize"
	"github.com/sky-as-code/nikki-erp/modules/contacts"
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/identity"
)

func LoadModules() ([]modules.NikkiModule, error) {
	return getStaticModules(), nil
}

func getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		// Sort alphabetically. The order of initialization will be handled properly.
		authorize.ModuleSingleton,
		authenticate.ModuleSingleton,
		contacts.ModuleSingleton,
		core.ModuleSingleton,
		identity.ModuleSingleton,
	}

	return modules
}
