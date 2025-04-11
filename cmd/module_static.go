//go:build !dynamicmods
// +build !dynamicmods

package main

import (
	"github.com/sky-as-code/nikki-erp/modules/core"
	"github.com/sky-as-code/nikki-erp/modules/shared"
	. "github.com/sky-as-code/nikki-erp/utility/fault"
)

func (thisApp *Application) getModules() ([]shared.NikkiModule, AppError) {
	return thisApp.getStaticModules(), nil
}

func (thisApp *Application) getStaticModules() []shared.NikkiModule {
	modules := []shared.NikkiModule{
		core.ModuleSingleton,
	}

	return modules
}
