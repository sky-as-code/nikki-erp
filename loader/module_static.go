//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/apptrait"
	"github.com/sky-as-code/nikki-erp/modules/authenticate"
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/essential"
	"github.com/sky-as-code/nikki-erp/modules/identity"
	"github.com/sky-as-code/nikki-erp/modules/inventory"
)

type StaticModuleLoader struct {
}

func (this StaticModuleLoader) LoadModules() ([]modules.InCodeModule, error) {
	return this.getStaticModules(), nil
}

func (this StaticModuleLoader) LoadModule(name string) (modules.InCodeModule, error) {
	allMods := this.getStaticModules()
	for _, mod := range allMods {
		if mod.Name() == name {
			return mod, nil
		}
	}
	return nil, errors.Errorf("module '%s' not found", name)
}

func (this StaticModuleLoader) getStaticModules() []modules.InCodeModule {
	modules := []modules.InCodeModule{
		// Sort alphabetically. The order of initialization will be handled properly.
		apptrait.ModuleSingleton,
		authenticate.ModuleSingleton,
		core.ModuleSingleton,
		// contacts.ModuleSingleton,
		essential.ModuleSingleton,
		identity.ModuleSingleton,
		inventory.ModuleSingleton,
	}

	return modules
}
