package external

import (
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itModule "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

// moduleCatalogAdapter groups Essential's dependency-ordered module list into waves.
type moduleCatalogAdapter struct {
	catalog itModule.ModuleCatalogService
}

func (this *moduleCatalogAdapter) ListModuleWaves(ctx corectx.Context) ([][]string, error) {
	mods, err := this.catalog.ListModulesByDeps(ctx)
	if err != nil {
		return nil, err
	}

	waves := make([][]string, 0)
	currentLevel := int32(-1)
	for _, mod := range mods {
		level := util.ValueOrZeroOf(mod.GetDepLevel())
		if len(waves) == 0 || level != currentLevel {
			waves = append(waves, []string{})
			currentLevel = level
		}
		waves[len(waves)-1] = append(waves[len(waves)-1], util.ValueOrZeroOf(mod.GetName()))
	}
	return waves, nil
}
