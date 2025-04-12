//go:build !dynamicmods
// +build !dynamicmods

package main

import (
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/shared"
	. "github.com/sky-as-code/nikki-erp/utility/fault"
)

func (thisApp *Application) getModules() ([]modules.NikkiModule, AppError) {
	return thisApp.getStaticModules(), nil
}

func (thisApp *Application) getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		shared.ModuleSingleton,
	}

	return modules
}
