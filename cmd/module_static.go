//go:build !dynamicmods
// +build !dynamicmods

package main

import (
	"github.com/sky-as-code/nikki-erp/common"
	. "github.com/sky-as-code/nikki-erp/common/util/fault"
	"github.com/sky-as-code/nikki-erp/modules"
)

func (thisApp *Application) getModules() ([]modules.NikkiModule, AppError) {
	return thisApp.getStaticModules(), nil
}

func (thisApp *Application) getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		common.ModuleSingleton,
	}

	return modules
}
