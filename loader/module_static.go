//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/common/module"
	"github.com/sky-as-code/nikki-erp/modules/authenticate"
	"github.com/sky-as-code/nikki-erp/modules/authorize"

	// "github.com/sky-as-code/nikki-erp/modules/contacts"
	"github.com/sky-as-code/nikki-erp/modules/essential"
	"github.com/sky-as-code/nikki-erp/modules/identity"
)

func LoadModules() ([]module.InCodeModule, error) {
	return getStaticModules(), nil
}

func getStaticModules() []module.InCodeModule {
	modules := []module.InCodeModule{
		// Sort alphabetically. The order of initialization will be handled properly.
		authorize.ModuleSingleton,
		authenticate.ModuleSingleton,
		// contacts.ModuleSingleton,
		essential.ModuleSingleton,
		identity.ModuleSingleton,
		// inventory.ModuleSingleton,
	}

	return modules
}
