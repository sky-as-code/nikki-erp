//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/authenticate"
	"github.com/sky-as-code/nikki-erp/modules/authorize"

	// "github.com/sky-as-code/nikki-erp/modules/contacts"
	"github.com/sky-as-code/nikki-erp/modules/essential"
	"github.com/sky-as-code/nikki-erp/modules/identity"
	"github.com/sky-as-code/nikki-erp/modules/drive"
)

func LoadModules() ([]modules.InCodeModule, error) {
	return getStaticModules(), nil
}

func getStaticModules() []modules.InCodeModule {
	modules := []modules.InCodeModule{
		// Sort alphabetically. The order of initialization will be handled properly.
		authorize.ModuleSingleton,
		authenticate.ModuleSingleton,
		// contacts.ModuleSingleton,
		essential.ModuleSingleton,
		identity.ModuleSingleton,
		// inventory.ModuleSingleton,
		drive.ModuleSingleton,
	}

	return modules
}
