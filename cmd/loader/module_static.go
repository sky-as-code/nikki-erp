//go:build !dynamicmods
// +build !dynamicmods

package loader

import (
	"github.com/sky-as-code/nikki-erp/common"
	. "github.com/sky-as-code/nikki-erp/common/util/fault"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core"
)

func LoadModules() ([]modules.NikkiModule, AppError) {
	return getStaticModules(), nil
}

func getStaticModules() []modules.NikkiModule {
	modules := []modules.NikkiModule{
		common.ModuleSingleton,
		core.ModuleSingleton,
	}

	return modules
}
