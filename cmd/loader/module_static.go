//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/contacts"
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/identity"
	"github.com/sky-as-code/nikki-erp/modules/authorize"
)

func LoadModules() ([]modules.NikkiModule, error) {
	return getStaticModules(), nil
}

func getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		// Sort alphabetically. The order of initialization will be handled properly.
		contacts.ModuleSingleton,
		core.ModuleSingleton,
		identity.ModuleSingleton,
		authorize.ModuleSingleton,
	}

	return modules
}
