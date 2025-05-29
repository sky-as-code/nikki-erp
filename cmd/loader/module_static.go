//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/identity"
)

func LoadModules() ([]modules.NikkiModule, error) {
	return getStaticModules(), nil
}

func getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		core.ModuleSingleton,
		identity.ModuleSingleton,
	}

	return modules
}
